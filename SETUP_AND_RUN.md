# NextState Go Backend - Setup & Run Guide

## Problem Diagnosis ❌

**Error encountered when running `go run main.go`:**
```
.\main.go:84:12: undefined: InitDB
.\main.go:87:8: undefined: DB
.\main.go:118:35: undefined: showLoginHandler
... (multiple undefined handler errors)
```

### Root Cause
When using `go run main.go`, **only the main.go file is compiled**. The other Go files containing handlers and database functions were not being included in the build:
- `db.go` - Contains `InitDB` and `DB` variable
- `auth.go` - Contains login/register/logout handlers
- `project.go` - Contains project handlers
- `task.go` - Contains task handlers
- `team.go` - Contains team handlers

---

## Solution ✅

### 1. Correct Way to Run (Option A - Direct Run)
```bash
cd BE/golang
go run .
```
The dot (`.`) tells Go to compile **all `.go` files** in the current directory.

### 2. Alternative: Build Executable (Option B)
```bash
cd BE/golang
go build .
```
This creates an executable file (`golang.exe` on Windows).

Then run:
```bash
./golang.exe  # Linux/Mac
golang.exe    # Windows
```

---

## Database Setup 🗄️

Before running the server, set up MySQL:

### Step 1: Start MySQL Service
```bash
# If using XAMPP
Start XAMPP Control Panel → MySQL → Start

# Or via command line
net start MySQL80
```

### Step 2: Create Database & Tables
```bash
# From the golang directory
mysql -u root -p < schema.sql

# Or enter MySQL CLI and paste schema.sql contents
mysql -u root
mysql> source schema.sql
```

### Step 3: Verify .env Configuration
File: `BE/golang/.env`
```env
PORT=8080
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=nextstate_db
DB_USERNAME=root
DB_PASSWORD=
SESSION_KEY=super-secret-nextstate-key-change-in-prod
```

**Important:** Adjust DB_PASSWORD if your MySQL root user has a password.

---

## Running the Server 🚀

### With Database Configured:
```bash
cd BE/golang
go run .
```

Expected output:
```
Database connection established successfully.
NextState Go server running at http://localhost:8080...
```

### Access the Application:
- **Login:** http://localhost:8080/login
- **Register:** http://localhost:8080/register
- **Dashboard:** http://localhost:8080/ (requires login)

---

## Troubleshooting 🔧

### "Connection refused" Error
```
failed to connect to database: dial tcp 127.0.0.1:3306: connectex: No connection 
could be made because the target machine actively refused it.
```
**Solution:** Start MySQL service

### "Unknown database" Error
```
Unknown database 'nextstate_db'
```
**Solution:** Run `mysql -u root < schema.sql` to create database

### "go command not found"
**Solution:** Install Go from https://golang.org/dl

---

## Project Structure
```
BE/golang/
├── main.go           # Server setup & routes
├── auth.go           # Login/Register/Logout handlers
├── project.go        # Project handlers
├── task.go           # Task handlers
├── team.go           # Team handlers
├── db.go             # Database models & queries
├── .env              # Configuration
├── schema.sql        # Database schema
├── go.mod            # Go module file
├── go.sum            # Dependency checksums
├── public/           # Static files
│   └── css/
├── templates/        # HTML templates
└── golang.exe        # Compiled executable (after build)
```

---

## Common Commands

| Command | Purpose |
|---------|---------|
| `go run .` | Run server directly |
| `go build .` | Create executable |
| `go fmt .` | Format code |
| `go vet .` | Check for errors |
| `go mod tidy` | Update dependencies |

---

**Version:** 1.0  
**Last Updated:** June 16, 2026
