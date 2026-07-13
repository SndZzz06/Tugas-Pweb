package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type CommonPageData struct {
	User            *User
	UserAvatar      string
	ActivePage      string
	MainContentFull bool
	CSRFToken       string
	Data            interface{}
}

// homeHandler displays the project lists (main dashboard)
func homeHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*User)
	projects, err := GetAllProjects()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load projects: %v", err), http.StatusInternalServerError)
		return
	}

	data := CommonPageData{
		User:            user,
		UserAvatar:      string(user.Name[0]),
		ActivePage:      "dashboard",
		MainContentFull: false,
		CSRFToken:       ensureCSRFToken(w, r),
		Data: map[string]interface{}{
			"Projects": projects,
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmplHomeView.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// projectCreateHandler shows project lists but with the creation modal overlay active
func projectCreateHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*User)
	projects, err := GetAllProjects()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load projects: %v", err), http.StatusInternalServerError)
		return
	}

	data := CommonPageData{
		User:            user,
		UserAvatar:      string(user.Name[0]),
		ActivePage:      "dashboard",
		MainContentFull: false,
		CSRFToken:       ensureCSRFToken(w, r),
		Data: map[string]interface{}{
			"Projects": projects,
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmplCreateProj.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// projectStoreHandler processes POST submit of new project
func projectStoreHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	color := r.FormValue("color")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")

	if name == "" || description == "" || color == "" || startDateStr == "" || endDateStr == "" {
		http.Error(w, "All fields are required.", http.StatusBadRequest)
		return
	}

	// Simple date ordering check
	startVal, err1 := time.Parse("2006-01-02", startDateStr)
	endVal, err2 := time.Parse("2006-01-02", endDateStr)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid date format.", http.StatusBadRequest)
		return
	}
	if endVal.Before(startVal) {
		http.Error(w, "End date must be after or equal to start date.", http.StatusBadRequest)
		return
	}

	proj := &Project{
		Name:        name,
		Description: description,
		Color:       color,
		StartDate:   startDateStr,
		EndDate:     endDateStr,
		Status:      "New",
	}

	_, err := CreateProject(proj)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create project: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// projectShowHandler shows details of a project, lists its tasks and handles filtering
func projectShowHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*User)
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	project, err := GetProjectByID(id)
	if err != nil || project == nil {
		http.NotFound(w, r)
		return
	}

	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	// Fetch all tasks for counting
	allTasks, err := GetTasksByProject(project.ID, "all")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	todoCount := 0
	inprogressCount := 0
	doneCount := 0
	for _, t := range allTasks {
		if t.Status == "To Do" {
			todoCount++
		} else if t.Status == "In Progress" {
			inprogressCount++
		} else if t.Status == "Done" {
			doneCount++
		}
	}

	// Fetch tasks matching filter
	var tasks []Task
	if filter == "all" {
		tasks = allTasks
	} else {
		tasks, err = GetTasksByProject(project.ID, filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Fetch users for assignment checklist
	teamUsers, err := GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := CommonPageData{
		User:            user,
		UserAvatar:      string(user.Name[0]),
		ActivePage:      "dashboard",
		MainContentFull: true,
		CSRFToken:       ensureCSRFToken(w, r),
		Data: map[string]interface{}{
			"Project":         project,
			"Tasks":           tasks,
			"Filter":          filter,
			"AllCount":        len(allTasks),
			"TodoCount":       todoCount,
			"InprogressCount": inprogressCount,
			"DoneCount":       doneCount,
			"Users":           teamUsers,
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmplShowProj.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// projectDestroyHandler handles DELETE /projects/{id}
func projectDestroyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = DeleteProject(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete project: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
