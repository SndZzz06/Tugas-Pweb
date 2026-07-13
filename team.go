package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// teamIndexHandler handles GET /team
func teamIndexHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*User)

	members, err := GetTeamMembers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load team members: %v", err), http.StatusInternalServerError)
		return
	}

	// Helper to format WhatsApp phone number (strip non-digits for links)
	reg := regexp.MustCompile("[^0-9]")
	membersData := make([]map[string]interface{}, len(members))
	for i, m := range members {
		membersData[i] = map[string]interface{}{
			"ID":          m.ID,
			"Name":        m.Name,
			"Role":        m.Role,
			"Phone":       m.Phone,
			"PhoneClean":  reg.ReplaceAllString(m.Phone, ""),
			"TotalTasks":  m.TotalTasks,
			"ActiveTasks": m.ActiveTasks,
			"LateTasks":   m.LateTasks,
		}
	}

	data := CommonPageData{
		User:            user,
		UserAvatar:      string(user.Name[0]),
		ActivePage:      "team",
		MainContentFull: true,
		CSRFToken:       ensureCSRFToken(w, r),
		Data: map[string]interface{}{
			"TeamMembers": membersData,
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmplTeamView.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// teamStoreHandler handles POST /team (adding a member)
func teamStoreHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	role := strings.TrimSpace(r.FormValue("role"))

	if name == "" || phone == "" || role == "" {
		http.Error(w, "Name, phone number, and role are required.", http.StatusBadRequest)
		return
	}

	// Generate random email like Laravel: slug + rand(1000, 9999) + @example.com
	rand.Seed(time.Now().UnixNano())
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	// remove non-alphanumeric chars
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	email := fmt.Sprintf("%s%d@example.com", slug, rand.Intn(9000)+1000)

	// Hash password: 'password123'
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to secure password", http.StatusInternalServerError)
		return
	}

	_, err = CreateUser(name, email, string(hashedPassword), phone, role)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add team member: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/team", http.StatusSeeOther)
}

// teamUpdateHandler handles PUT /team/{id}
func teamUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	role := strings.TrimSpace(r.FormValue("role"))

	if name == "" || phone == "" || role == "" {
		http.Error(w, "Name, phone number, and role are required.", http.StatusBadRequest)
		return
	}

	err = UpdateUser(id, name, phone, role)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update team member: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/team", http.StatusSeeOther)
}

// teamDestroyHandler handles DELETE /team/{id}
func teamDestroyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = DeleteUser(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete team member: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/team", http.StatusSeeOther)
}
