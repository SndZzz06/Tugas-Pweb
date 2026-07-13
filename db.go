package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// Models
type User struct {
	ID        int64
	Name      string
	Email     string // nullable in migration, but we treat it as string
	Phone     string
	Password  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Project struct {
	ID          int64
	Name        string
	Description string
	Color       string
	StartDate   string // YYYY-MM-DD format
	EndDate     string // YYYY-MM-DD format
	Status      string // New, Active, Closed
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Task struct {
	ID          int64
	ProjectID   int64
	Title       string
	Description string
	StartDate   string // YYYY-MM-DD format, nullable
	DueDate     string // YYYY-MM-DD format, nullable
	Priority    string // Easy, Medium, High
	Status      string // To Do, In Progress, Done
	Attachment  string // nullable
	Assignees   []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TeamMember struct {
	ID          int64
	Name        string
	Role        string
	Phone       string
	TotalTasks  int
	ActiveTasks int
	LateTasks   int
}

// InitDB initializes database connection
func InitDB(dsn string) error {
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// Wait for connection to be ready (up to 10 seconds)
	for i := 0; i < 10; i++ {
		err = DB.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for DB connection... (%d/10)", i+1)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Println("Database connection established successfully.")
	return nil
}

// User helper methods

func GetUserByEmail(email string) (*User, error) {
	var u User
	var emailNull, phoneNull, roleNull sql.NullString
	query := "SELECT id, name, email, phone, password, role, created_at, updated_at FROM users WHERE email = ?"
	err := DB.QueryRow(query, email).Scan(&u.ID, &u.Name, &emailNull, &phoneNull, &u.Password, &roleNull, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.Email = emailNull.String
	u.Phone = phoneNull.String
	u.Role = roleNull.String
	return &u, nil
}

func GetUserByID(id int64) (*User, error) {
	var u User
	var emailNull, phoneNull, roleNull sql.NullString
	query := "SELECT id, name, email, phone, password, role, created_at, updated_at FROM users WHERE id = ?"
	err := DB.QueryRow(query, id).Scan(&u.ID, &u.Name, &emailNull, &phoneNull, &u.Password, &roleNull, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.Email = emailNull.String
	u.Phone = phoneNull.String
	u.Role = roleNull.String
	return &u, nil
}

func GetAllUsers() ([]User, error) {
	query := "SELECT id, name, email, phone, role FROM users"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var emailNull, phoneNull, roleNull sql.NullString
		if err := rows.Scan(&u.ID, &u.Name, &emailNull, &phoneNull, &roleNull); err != nil {
			return nil, err
		}
		u.Email = emailNull.String
		u.Phone = phoneNull.String
		u.Role = roleNull.String
		users = append(users, u)
	}
	return users, nil
}

func CreateUser(name, email, password, phone, role string) (int64, error) {
	query := "INSERT INTO users (name, email, password, phone, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW())"
	// Set default values if empty
	var emailVal, phoneVal, roleVal interface{}
	emailVal = email
	if email == "" {
		emailVal = nil
	}
	phoneVal = phone
	if phone == "" {
		phoneVal = nil
	}
	roleVal = role
	if role == "" {
		roleVal = "Member"
	}

	res, err := DB.Exec(query, name, emailVal, password, phoneVal, roleVal)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateUser(id int64, name, phone, role string) error {
	query := "UPDATE users SET name = ?, phone = ?, role = ?, updated_at = NOW() WHERE id = ?"
	_, err := DB.Exec(query, name, phone, role, id)
	return err
}

func DeleteUser(id int64) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := DB.Exec(query, id)
	return err
}

// Project helper methods

func GetAllProjects() ([]Project, error) {
	query := "SELECT id, name, description, color, start_date, end_date, status, created_at, updated_at FROM projects ORDER BY created_at DESC"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		var descNull, colorNull, statusNull sql.NullString
		var startDateBytes, endDateBytes []byte // date fields scan cleanly as []byte or string
		if err := rows.Scan(&p.ID, &p.Name, &descNull, &colorNull, &startDateBytes, &endDateBytes, &statusNull, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Description = descNull.String
		p.Color = colorNull.String
		p.Status = statusNull.String
		p.StartDate = string(startDateBytes)
		p.EndDate = string(endDateBytes)
		projects = append(projects, p)
	}
	return projects, nil
}

func GetProjectByID(id int64) (*Project, error) {
	var p Project
	var descNull, colorNull, statusNull sql.NullString
	var startDateBytes, endDateBytes []byte
	query := "SELECT id, name, description, color, start_date, end_date, status, created_at, updated_at FROM projects WHERE id = ?"
	err := DB.QueryRow(query, id).Scan(&p.ID, &p.Name, &descNull, &colorNull, &startDateBytes, &endDateBytes, &statusNull, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Description = descNull.String
	p.Color = colorNull.String
	p.Status = statusNull.String
	p.StartDate = string(startDateBytes)
	p.EndDate = string(endDateBytes)
	return &p, nil
}

func CreateProject(p *Project) (int64, error) {
	query := "INSERT INTO projects (name, description, color, start_date, end_date, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())"
	res, err := DB.Exec(query, p.Name, p.Description, p.Color, p.StartDate, p.EndDate, p.Status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeleteProject(id int64) error {
	query := "DELETE FROM projects WHERE id = ?"
	_, err := DB.Exec(query, id)
	return err
}

// Task helper methods

func GetTasksByProject(projectID int64, filter string) ([]Task, error) {
	var query string
	var args []interface{}

	if filter == "to-do" {
		query = "SELECT id, project_id, title, description, start_date, due_date, priority, status, attachment, created_at, updated_at FROM tasks WHERE project_id = ? AND status = 'To Do' ORDER BY created_at DESC"
		args = append(args, projectID)
	} else if filter == "in-progress" {
		query = "SELECT id, project_id, title, description, start_date, due_date, priority, status, attachment, created_at, updated_at FROM tasks WHERE project_id = ? AND status = 'In Progress' ORDER BY created_at DESC"
		args = append(args, projectID)
	} else if filter == "done" {
		query = "SELECT id, project_id, title, description, start_date, due_date, priority, status, attachment, created_at, updated_at FROM tasks WHERE project_id = ? AND status = 'Done' ORDER BY created_at DESC"
		args = append(args, projectID)
	} else {
		query = "SELECT id, project_id, title, description, start_date, due_date, priority, status, attachment, created_at, updated_at FROM tasks WHERE project_id = ? ORDER BY created_at DESC"
		args = append(args, projectID)
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var descNull, priorityNull, statusNull, attachNull sql.NullString
		var startBytes, dueBytes []byte
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &descNull, &startBytes, &dueBytes, &priorityNull, &statusNull, &attachNull, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Description = descNull.String
		t.Priority = priorityNull.String
		t.Status = statusNull.String
		t.Attachment = attachNull.String
		t.StartDate = string(startBytes)
		t.DueDate = string(dueBytes)

		// Get assignees names
		assignees, err := GetTaskAssignees(t.ID)
		if err != nil {
			return nil, err
		}
		t.Assignees = assignees

		tasks = append(tasks, t)
	}
	return tasks, nil
}

func GetTaskAssignees(taskID int64) ([]string, error) {
	query := "SELECT u.name FROM users u JOIN task_user tu ON u.id = tu.user_id WHERE tu.task_id = ?"
	rows, err := DB.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func GetTaskByID(id int64) (*Task, error) {
	var t Task
	var descNull, priorityNull, statusNull, attachNull sql.NullString
	var startBytes, dueBytes []byte
	query := "SELECT id, project_id, title, description, start_date, due_date, priority, status, attachment, created_at, updated_at FROM tasks WHERE id = ?"
	err := DB.QueryRow(query, id).Scan(&t.ID, &t.ProjectID, &t.Title, &descNull, &startBytes, &dueBytes, &priorityNull, &statusNull, &attachNull, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.Description = descNull.String
	t.Priority = priorityNull.String
	t.Status = statusNull.String
	t.Attachment = attachNull.String
	t.StartDate = string(startBytes)
	t.DueDate = string(dueBytes)
	return &t, nil
}

func CreateTask(t *Task, assignees []int64) (int64, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := "INSERT INTO tasks (project_id, title, description, start_date, due_date, priority, status, attachment, created_at, updated_at) VALUES (?, ?, ?, NOW(), ?, ?, ?, ?, NOW(), NOW())"
	var dueVal interface{} = t.DueDate
	if t.DueDate == "" {
		dueVal = nil
	}
	var attachVal interface{} = t.Attachment
	if t.Attachment == "" {
		attachVal = nil
	}

	res, err := tx.Exec(query, t.ProjectID, t.Title, t.Description, dueVal, t.Priority, t.Status, attachVal)
	if err != nil {
		return 0, err
	}

	taskID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert assignees in pivot table
	for _, userID := range assignees {
		_, err := tx.Exec("INSERT INTO task_user (task_id, user_id, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", taskID, userID)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return taskID, nil
}

func UpdateTask(t *Task) error {
	var query string
	var args []interface{}

	if t.Attachment != "" {
		query = "UPDATE tasks SET title = ?, description = ?, due_date = ?, priority = ?, status = ?, attachment = ?, updated_at = NOW() WHERE id = ?"
		var dueVal interface{} = t.DueDate
		if t.DueDate == "" {
			dueVal = nil
		}
		args = append(args, t.Title, t.Description, dueVal, t.Priority, t.Status, t.Attachment, t.ID)
	} else {
		query = "UPDATE tasks SET title = ?, description = ?, due_date = ?, priority = ?, status = ?, updated_at = NOW() WHERE id = ?"
		var dueVal interface{} = t.DueDate
		if t.DueDate == "" {
			dueVal = nil
		}
		args = append(args, t.Title, t.Description, dueVal, t.Priority, t.Status, t.ID)
	}

	_, err := DB.Exec(query, args...)
	return err
}

func DeleteTask(id int64) error {
	query := "DELETE FROM tasks WHERE id = ?"
	_, err := DB.Exec(query, id)
	return err
}

// Team management queries

func GetTeamMembers() ([]TeamMember, error) {
	// Replicating Laravel's tasks, active_tasks and late_tasks counts
	// active_tasks: status in ('To Do', 'In Progress')
	// late_tasks: due_date < NOW() AND status != 'Done'
	query := `
		SELECT 
			u.id, 
			u.name, 
			COALESCE(u.role, 'Member') as role, 
			COALESCE(u.phone, '+6280000000000') as phone,
			(SELECT COUNT(*) FROM task_user tu JOIN tasks t ON tu.task_id = t.id WHERE tu.user_id = u.id) as total_tasks,
			(SELECT COUNT(*) FROM task_user tu JOIN tasks t ON tu.task_id = t.id WHERE tu.user_id = u.id AND t.status IN ('To Do', 'In Progress')) as active_tasks,
			(SELECT COUNT(*) FROM task_user tu JOIN tasks t ON tu.task_id = t.id WHERE tu.user_id = u.id AND t.due_date < NOW() AND t.status != 'Done') as late_tasks
		FROM users u
		ORDER BY u.name ASC
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []TeamMember
	for rows.Next() {
		var m TeamMember
		if err := rows.Scan(&m.ID, &m.Name, &m.Role, &m.Phone, &m.TotalTasks, &m.ActiveTasks, &m.LateTasks); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}
