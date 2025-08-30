@echo off
echo === Restarting backend user_management ===

REM Check if there are any go run processes still running on port 8081
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8081') do (
    echo Killing process with PID %%a
    taskkill /PID %%a /F >nul 2>&1
)

REM Change to project directory
cd /d "C:\Users\prima\SynologyDrive\golang\lokstra\cmd\projects\user_management"

REM Start the backend again
start cmd /k "go run ."