package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// taskStoreHandler handles POST /projects/{project}/tasks
func taskStoreHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectIDStr := vars["project"]

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Parse multipart form (up to 32MB)
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	status := r.FormValue("status")
	priority := r.FormValue("priority")
	dueDate := r.FormValue("due_date")

	if title == "" || status == "" || priority == "" {
		http.Error(w, "Title, status, and priority are required.", http.StatusBadRequest)
		return
	}

	// Retrieve assignees (multiple checkboxes)
	assigneesStrList := r.Form["assignees[]"]
	var assignees []int64
	for _, idStr := range assigneesStrList {
		uid, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			assignees = append(assignees, uid)
		}
	}

	// Handle attachment upload
	var attachmentPath string
	file, header, err := r.FormFile("attachment")
	if err == nil {
		defer file.Close()

		// Ensure target directory exists
		uploadDir := "./public/storage/attachments"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
			return
		}

		// Generate unique name
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
		destPath := filepath.Join(uploadDir, filename)

		out, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "Failed to save upload file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Failed to write upload file", http.StatusInternalServerError)
			return
		}

		// Save relative path matching Laravel's disk structure
		attachmentPath = "attachments/" + filename
	}

	task := &Task{
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      status,
		Priority:    priority,
		DueDate:     dueDate,
		Attachment:  attachmentPath,
	}

	_, err = CreateTask(task, assignees)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to store task: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects/"+projectIDStr, http.StatusSeeOther)
}

// taskUpdateHandler handles PUT /tasks/{id}
func taskUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Retrieve original task to fetch its project ID
	existingTask, err := GetTaskByID(id)
	if err != nil || existingTask == nil {
		http.NotFound(w, r)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	status := r.FormValue("status")
	priority := r.FormValue("priority")
	dueDate := r.FormValue("due_date")

	if title == "" || status == "" || priority == "" {
		http.Error(w, "Title, status, and priority are required.", http.StatusBadRequest)
		return
	}

	// Handle optional attachment upload
	var attachmentPath string
	file, header, err := r.FormFile("attachment")
	if err == nil {
		defer file.Close()

		uploadDir := "./public/storage/attachments"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
		destPath := filepath.Join(uploadDir, filename)

		out, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "Failed to save upload file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Failed to write upload file", http.StatusInternalServerError)
			return
		}

		attachmentPath = "attachments/" + filename
	}

	task := &Task{
		ID:          id,
		ProjectID:   existingTask.ProjectID,
		Title:       title,
		Description: description,
		Status:      status,
		Priority:    priority,
		DueDate:     dueDate,
		Attachment:  attachmentPath, // Only updated in db if not empty
	}

	err = UpdateTask(task)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update task: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect back to referer or project page
	referer := r.Header.Get("Referer")
	if referer != "" {
		http.Redirect(w, r, referer, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/projects/"+strconv.FormatInt(existingTask.ProjectID, 10), http.StatusSeeOther)
	}
}

// taskDestroyHandler handles DELETE /tasks/{id}
func taskDestroyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	existingTask, err := GetTaskByID(id)
	if err != nil || existingTask == nil {
		http.NotFound(w, r)
		return
	}

	err = DeleteTask(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete task: %v", err), http.StatusInternalServerError)
		return
	}

	referer := r.Header.Get("Referer")
	if referer != "" {
		http.Redirect(w, r, referer, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/projects/"+strconv.FormatInt(existingTask.ProjectID, 10), http.StatusSeeOther)
	}
}
