@echo off
REM Setup NextState Database
echo ============================================
echo NextState Database Setup
echo ============================================
echo.
echo This script will:
echo 1. Create database and tables
echo 2. Verify MySQL connection
echo.

REM Check if MySQL is accessible
mysql -u root -e "SELECT 1" >nul 2>&1
if errorlevel 1 (
    echo ERROR: MySQL is not running or not accessible!
    echo.
    echo SOLUTION:
    echo 1. Open XAMPP Control Panel
    echo 2. Click "Start" button next to MySQL
    echo 3. Wait for it to start (status should be "Running")
    echo 4. Run this script again
    echo.
    pause
    exit /b 1
)

echo MySQL is running! Creating database and tables...
echo.

REM Create database and tables
mysql -u root < schema.sql

if errorlevel 1 (
    echo ERROR: Failed to create database!
    pause
    exit /b 1
)

echo.
echo ============================================
echo SUCCESS! Database setup complete!
echo ============================================
echo.
echo You can now run the server with: go run .
echo.
pause
