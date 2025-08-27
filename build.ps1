# Cross-platform build script for Lokstra (PowerShell)

param(
    [string]$Version = "1.0.0",
    [string]$AppName = "lokstra"
)

$BuildDir = "./build"
$ErrorActionPreference = "Stop"

# Create build directory
if (!(Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir | Out-Null
}

Write-Host "Building Lokstra v$Version for multiple platforms..." -ForegroundColor Green

# Windows AMD64
Write-Host "Building for Windows AMD64..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags="-s -w" -o "$BuildDir/$AppName-$Version-windows-amd64.exe" .

# Linux AMD64
Write-Host "Building for Linux AMD64..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags="-s -w" -o "$BuildDir/$AppName-$Version-linux-amd64" .

# Linux ARM64
Write-Host "Building for Linux ARM64..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "arm64"
go build -ldflags="-s -w" -o "$BuildDir/$AppName-$Version-linux-arm64" .

# macOS AMD64
Write-Host "Building for macOS AMD64..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -ldflags="-s -w" -o "$BuildDir/$AppName-$Version-darwin-amd64" .

# macOS ARM64 (Apple Silicon)
Write-Host "Building for macOS ARM64..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -ldflags="-s -w" -o "$BuildDir/$AppName-$Version-darwin-arm64" .

Write-Host "✓ All builds completed successfully!" -ForegroundColor Green
Write-Host "Build artifacts are in: $BuildDir" -ForegroundColor Cyan
Get-ChildItem $BuildDir | Format-Table Name, Length, LastWriteTime
