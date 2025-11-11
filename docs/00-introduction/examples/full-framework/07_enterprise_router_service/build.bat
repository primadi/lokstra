@echo off
REM Build script for Windows CMD - ensures code generation before build
REM Usage: 
REM   build.bat           - Build for current platform (Windows)
REM   build.bat windows   - Build for Windows
REM   build.bat linux     - Build for Linux
REM   build.bat darwin    - Build for macOS

setlocal

set TARGET_OS=%1
if "%TARGET_OS%"=="" set TARGET_OS=windows

echo.
echo ╔════════════════════════════════════════════════════╗
echo ║  Lokstra Build Script - Code Gen + Build          ║
echo ╚════════════════════════════════════════════════════╝
echo.

REM Step 1: Generate code
echo Step 1/4: Generating code (forced rebuild)...
echo   Running 'go run . --generate-only'...

go run . --generate-only
if errorlevel 1 (
    echo   ✗ Code generation failed!
    exit /b 1
)
echo   ✓ Code generation completed
echo.

REM Step 2: Tidy dependencies
echo Step 2/4: Tidying dependencies...
echo   Running 'go mod tidy'...

go mod tidy >nul 2>&1
if errorlevel 1 (
    echo   ⚠ Warning: go mod tidy failed ^(continuing anyway^)
) else (
    echo   ✓ Dependencies tidied
)
echo.

REM Step 3: Build binary
echo Step 3/4: Building binary for %TARGET_OS%...

if /i "%TARGET_OS%"=="linux" (
    echo   Building for Linux...
    set BINARY_NAME=app-linux
    set GOOS=linux
    set GOARCH=amd64
    go build -o app-linux .
    if errorlevel 1 (
        echo   ✗ Build failed!
        exit /b 1
    )
    echo   ✓ Build successful: app-linux
) else if /i "%TARGET_OS%"=="windows" (
    echo   Building for Windows...
    set BINARY_NAME=app-windows.exe
    set GOOS=windows
    set GOARCH=amd64
    go build -o app-windows.exe .
    if errorlevel 1 (
        echo   ✗ Build failed!
        exit /b 1
    )
    echo   ✓ Build successful: app-windows.exe
) else if /i "%TARGET_OS%"=="darwin" (
    echo   Building for macOS...
    set BINARY_NAME=app-darwin
    set GOOS=darwin
    set GOARCH=amd64
    go build -o app-darwin .
    if errorlevel 1 (
        echo   ✗ Build failed!
        exit /b 1
    )
    echo   ✓ Build successful: app-darwin
) else (
    echo   Building for current platform...
    set BINARY_NAME=app.exe
    go build -o app.exe .
    if errorlevel 1 (
        echo   ✗ Build failed!
        exit /b 1
    )
    echo   ✓ Build successful: app.exe
)

echo.
echo ╔════════════════════════════════════════════════════╗
echo ║  Build Complete!                                   ║
echo ╚════════════════════════════════════════════════════╝
echo.
echo Binary created: %BINARY_NAME%
echo.
echo Usage:
echo   build.bat           - Build for current platform (Windows)
echo   build.bat windows   - Build for Windows
echo   build.bat linux     - Build for Linux
echo   build.bat darwin    - Build for macOS
echo.

endlocal
