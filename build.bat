@echo off
echo Building NextState Go Backend...
echo.
go build -o golang.exe .
if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build successful! Run: golang.exe
) else (
    echo Build failed!
)
pause
