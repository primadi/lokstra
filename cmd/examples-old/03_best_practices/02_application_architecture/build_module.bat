@echo off
REM Build script for user management module on Windows
REM This script compiles the user management module as a shared library (.dll/.so file)

echo Building user management module...

REM Set build flags
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

REM Build the module as a shared library
go build -buildmode=plugin -o user_management.so .\modules\user_management\

if %ERRORLEVEL% equ 0 (
    echo ✅ User management module built successfully: user_management.so
    
    REM Show file information
    echo 📁 Plugin file information:
    dir user_management.so
) else (
    echo ❌ Failed to build user management module
    exit /b 1
)

echo.
echo 📝 Usage:
echo   1. Update lokstra.yaml to reference the built plugin:
echo      modules:
echo        - name: "user_management"
echo          path: "./user_management.so"
echo          entry: "RegisterModule"
echo.
echo   2. Run your application:
echo      go run main.go
