@echo off
echo ============================================
echo NextState Backend - Quick Start
echo ============================================
echo.
echo Step 1: Ensure MySQL is running
echo   - Open XAMPP Control Panel
echo   - Click "Start" next to MySQL
echo   - Wait until it says "Running"
echo.
echo Step 2: Setup Database (first time only)
echo   - Double-click: setup_db.bat
echo.
echo Step 3: Run the Server
echo   - Double-click: run.bat
echo   OR in terminal: go run .
echo.
echo Step 4: Access Application
echo   - Open browser: http://localhost:8080
echo.
echo ============================================
echo Commands Reference:
echo ============================================
echo go run .           - Run server
echo go build .         - Build executable
echo go fmt .           - Format code
echo go vet .           - Check for errors
echo.
pause
