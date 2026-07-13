package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var sessionStore *sessions.CookieStore
var (
	tmplLogin      *template.Template
	tmplRegister   *template.Template
	tmplHomeView   *template.Template
	tmplCreateProj *template.Template
	tmplShowProj   *template.Template
	tmplTeamView   *template.Template
)

type rateInfo struct {
	count       int
	windowStart time.Time
}

type loginAttempt struct {
	count       int
	lastAttempt time.Time
	lockUntil   time.Time
}

var (
	requestRates    = make(map[string]*rateInfo)
	requestRatesMu  sync.Mutex
	loginAttempts   = make(map[string]*loginAttempt)
	loginAttemptsMu sync.Mutex
)

// Context keys
type contextKey string

const userContextKey contextKey = "user"

func initTemplates() {
	// Parse individual files. Go's html/template lets us compile helper functions if needed.
	funcMap := template.FuncMap{
		"tolower": strings.ToLower,
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"implode": func(sep string, items []string) string {
			return strings.Join(items, sep)
		},
		"avatar": func(name string) string {
			if len(name) == 0 {
				return "?"
			}
			return strings.ToUpper(string(name[0]))
		},
		"json": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
	}

	tmplLogin = template.Must(template.New("login.html").Funcs(funcMap).ParseFiles("templates/login.html"))
	tmplRegister = template.Must(template.New("register.html").Funcs(funcMap).ParseFiles("templates/register.html"))

	tmplHomeView = template.Must(template.New("layout").Funcs(funcMap).ParseFiles("templates/layout.html", "templates/home.html"))
	tmplCreateProj = template.Must(template.New("layout").Funcs(funcMap).ParseFiles("templates/layout.html", "templates/create_project.html"))
	tmplShowProj = template.Must(template.New("layout").Funcs(funcMap).ParseFiles("templates/layout.html", "templates/show_project.html"))
	tmplTeamView = template.Must(template.New("layout").Funcs(funcMap).ParseFiles("templates/layout.html", "templates/team.html"))
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Database connection DSN
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_DATABASE")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
	if err := InitDB(dsn); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer DB.Close()

	// Sessions store initialization
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		sessionKey = "default-nextstate-session-key"
	}
	sessionStore = sessions.NewCookieStore([]byte(sessionKey))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	// Initialize template configurations
	initTemplates()

	// Initialize Gorilla Mux Router
	r := mux.NewRouter()

	// Middlewares
	r.Use(loggingMiddleware)
	r.Use(securityHeadersMiddleware)
	r.Use(rateLimitMiddleware)
	r.Use(csrfMiddleware)

	// Static files handling
	r.PathPrefix("/css/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./public"))))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./public"))))
	// Upload attachments serving
	r.PathPrefix("/storage/").Handler(http.StripPrefix("/storage/", http.FileServer(http.Dir("./public/storage"))))

	// Guest routes (redirect to home if authenticated)
	guestRouter := r.NewRoute().Subrouter()
	guestRouter.Use(guestMiddleware)
	guestRouter.HandleFunc("/login", showLoginHandler).Methods("GET")
	guestRouter.HandleFunc("/login", loginHandler).Methods("POST")
	guestRouter.HandleFunc("/register", showRegisterHandler).Methods("GET")
	guestRouter.HandleFunc("/register", registerHandler).Methods("POST")

	// Authenticated routes
	authRouter := r.NewRoute().Subrouter()
	authRouter.Use(authMiddleware)

	// Home
	authRouter.HandleFunc("/", homeHandler).Methods("GET")

	// Projects
	authRouter.HandleFunc("/projects/create", projectCreateHandler).Methods("GET")
	authRouter.HandleFunc("/projects", projectStoreHandler).Methods("POST")
	authRouter.HandleFunc("/projects/{id}", projectShowHandler).Methods("GET")
	authRouter.HandleFunc("/projects/{id}", projectDestroyHandler).Methods("DELETE")

	// Tasks
	authRouter.HandleFunc("/projects/{project}/tasks", taskStoreHandler).Methods("POST")
	authRouter.HandleFunc("/tasks/{id}", taskUpdateHandler).Methods("PUT")
	authRouter.HandleFunc("/tasks/{id}", taskDestroyHandler).Methods("DELETE")

	// Team
	authRouter.HandleFunc("/team", teamIndexHandler).Methods("GET")
	authRouter.HandleFunc("/team", teamStoreHandler).Methods("POST")
	authRouter.HandleFunc("/team/{id}", teamUpdateHandler).Methods("PUT")
	authRouter.HandleFunc("/team/{id}", teamDestroyHandler).Methods("DELETE")

	// Logout
	authRouter.HandleFunc("/logout", logoutHandler).Methods("POST")

	// Setup Server
	addr := ":" + port
	fmt.Printf("NextState Go server running at http://localhost%s...\n", addr)

	// Wrap router with methodOverrideMiddleware so that it rewrites PUT/DELETE before routing
	err := http.ListenAndServe(addr, methodOverrideMiddleware(r))
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Global Middlewares

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s - %v",
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}

func methodOverrideMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Parse multipart form up to 32MB to extract form values
			r.ParseMultipartForm(32 << 20)
			if method := r.FormValue("_method"); method != "" {
				methodUpper := strings.ToUpper(method)
				if methodUpper == "PUT" || methodUpper == "DELETE" || methodUpper == "PATCH" {
					r.Method = methodUpper
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, "nextstate-session")
		uidVal := session.Values["user_id"]
		if uidVal == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var userID int64
		switch v := uidVal.(type) {
		case int64:
			userID = v
		case int:
			userID = int64(v)
		case string:
			id, _ := strconv.ParseInt(v, 10, 64)
			userID = id
		default:
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := GetUserByID(userID)
		if err != nil || user == nil {
			// Invalidate session
			session.Options.MaxAge = -1
			session.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Inject user into context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func guestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, "nextstate-session")
		if session.Values["user_id"] != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func generateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func ensureCSRFToken(w http.ResponseWriter, r *http.Request) string {
	session, _ := sessionStore.Get(r, "nextstate-session")
	token, _ := session.Values["csrf_token"].(string)
	if token == "" {
		newToken, err := generateRandomToken(32)
		if err != nil {
			return ""
		}
		token = newToken
		session.Values["csrf_token"] = token
		session.Save(r, w)
	}
	return token
}

func validateCSRFToken(r *http.Request) bool {
	session, _ := sessionStore.Get(r, "nextstate-session")
	token, _ := session.Values["csrf_token"].(string)
	if token == "" {
		return false
	}
	csfrToken := r.FormValue("csrf_token")
	return csfrToken != "" && csfrToken == token
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com; font-src 'self' https://fonts.gstatic.com https://cdnjs.cloudflare.com; img-src 'self' data:;")
		next.ServeHTTP(w, r)
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		requestRatesMu.Lock()
		rate, ok := requestRates[ip]
		if !ok || time.Since(rate.windowStart) > time.Minute {
			rate = &rateInfo{count: 0, windowStart: time.Now()}
			requestRates[ip] = rate
		}
		rate.count++
		count := rate.count
		requestRatesMu.Unlock()

		if count > 120 {
			http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete || r.Method == http.MethodPatch {
			// Ensure form values are available for CSRF validation
			r.ParseMultipartForm(32 << 20)
			if !validateCSRFToken(r) {
				http.Error(w, "Invalid CSRF token.", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
