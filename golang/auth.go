package main

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthPageData struct {
	Errors    []string
	Name      string
	Email     string
	CSRFToken string
}

func getLoginAttemptState(ip string) (*loginAttempt, bool) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	attempt, ok := loginAttempts[ip]
	if !ok {
		return nil, false
	}
	if time.Now().Before(attempt.lockUntil) {
		return attempt, true
	}
	return attempt, false
}

func incrementLoginAttempt(ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	attempt, ok := loginAttempts[ip]
	if !ok {
		attempt = &loginAttempt{count: 0, lastAttempt: time.Now()}
		loginAttempts[ip] = attempt
	}
	attempt.count++
	attempt.lastAttempt = time.Now()
	if attempt.count >= 5 {
		attempt.lockUntil = time.Now().Add(15 * time.Minute)
	}
}

func clearLoginAttempts(ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	delete(loginAttempts, ip)
}

func showLoginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "nextstate-session")
	flashes := session.Flashes("errors")
	session.Save(r, w)

	var errors []string
	for _, f := range flashes {
		if str, ok := f.(string); ok {
			errors = append(errors, str)
		}
	}

	data := AuthPageData{
		Errors:    errors,
		CSRFToken: ensureCSRFToken(w, r),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmplLogin.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !validateCSRFToken(r) {
		http.Error(w, "Invalid CSRF token.", http.StatusForbidden)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	ip := getClientIP(r)

	if _, locked := getLoginAttemptState(ip); locked {
		saveFlashErrorsAndRedirect(w, r, []string{"Too many login attempts. Silakan hubungi administrator."}, "/login")
		return
	}

	var errors []string
	if email == "" {
		errors = append(errors, "Email is required.")
	}
	if password == "" {
		errors = append(errors, "Password is required.")
	}

	if len(errors) > 0 {
		incrementLoginAttempt(ip)
		saveFlashErrorsAndRedirect(w, r, errors, "/login")
		return
	}

	user, err := GetUserByEmail(email)
	if err != nil {
		incrementLoginAttempt(ip)
		errors = append(errors, "Database query error.")
		saveFlashErrorsAndRedirect(w, r, errors, "/login")
		return
	}

	if user == nil {
		incrementLoginAttempt(ip)
		errors = append(errors, "The provided credentials do not match our records.")
		saveFlashErrorsAndRedirect(w, r, errors, "/login")
		return
	}

	// Compare bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		incrementLoginAttempt(ip)
		errors = append(errors, "The provided credentials do not match our records.")
		saveFlashErrorsAndRedirect(w, r, errors, "/login")
		return
	}

	clearLoginAttempts(ip)

	// Session login
	session, _ := sessionStore.Get(r, "nextstate-session")
	session.Values["user_id"] = user.ID

	// Handle remember me checkbox
	remember := r.FormValue("remember")
	if remember == "on" {
		session.Options.MaxAge = 3600 * 24 * 30 // 30 days
	} else {
		session.Options.MaxAge = 3600 * 2 // 2 hours
	}

	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showRegisterHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "nextstate-session")
	flashes := session.Flashes("errors")
	session.Save(r, w)

	var errors []string
	for _, f := range flashes {
		if str, ok := f.(string); ok {
			errors = append(errors, str)
		}
	}

	data := AuthPageData{
		Errors:    errors,
		CSRFToken: ensureCSRFToken(w, r),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmplRegister.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	passwordConf := r.FormValue("password_confirmation")

	var errors []string
	if name == "" {
		errors = append(errors, "Name is required.")
	}
	if email == "" {
		errors = append(errors, "Email is required.")
	}
	if len(password) < 6 {
		errors = append(errors, "Password must be at least 6 characters.")
	}
	if password != passwordConf {
		errors = append(errors, "Password confirmation does not match.")
	}

	if len(errors) > 0 {
		saveFlashErrorsAndRedirect(w, r, errors, "/register")
		return
	}

	// Check email uniqueness
	existingUser, err := GetUserByEmail(email)
	if err != nil {
		errors = append(errors, "Database query error.")
		saveFlashErrorsAndRedirect(w, r, errors, "/register")
		return
	}
	if existingUser != nil {
		errors = append(errors, "The email address is already registered.")
		saveFlashErrorsAndRedirect(w, r, errors, "/register")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		errors = append(errors, "Error securing password.")
		saveFlashErrorsAndRedirect(w, r, errors, "/register")
		return
	}

	// Default role as Admin if it's the first user, otherwise Member
	role := "Member"
	allUsers, _ := GetAllUsers()
	if len(allUsers) == 0 {
		role = "Admin"
	}

	userID, err := CreateUser(name, email, string(hashedPassword), "", role)
	if err != nil {
		errors = append(errors, "Failed to create user account.")
		saveFlashErrorsAndRedirect(w, r, errors, "/register")
		return
	}

	// Auto-login
	session, _ := sessionStore.Get(r, "nextstate-session")
	session.Values["user_id"] = userID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "nextstate-session")
	delete(session.Values, "user_id")
	session.Options.MaxAge = -1 // delete cookie immediately
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func saveFlashErrorsAndRedirect(w http.ResponseWriter, r *http.Request, errors []string, redirectPath string) {
	session, _ := sessionStore.Get(r, "nextstate-session")
	for _, e := range errors {
		session.AddFlash(e, "errors")
	}
	session.Save(r, w)
	http.Redirect(w, r, redirectPath, http.StatusSeeOther)
}
