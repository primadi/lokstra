# Build script for Windows - ensures code generation before build
# Usage: 
#   .\build.ps1           - Build for current platform (Windows)
#   .\build.ps1 windows   - Build for Windows
#   .\build.ps1 linux     - Build for Linux
#   .\build.ps1 darwin    - Build for macOS

param(
    [string]$TargetOS = "windows"
)

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  Lokstra Build Script - Code Gen + Build          ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Step 1: Generate code
Write-Host "Step 1/4: Generating code (forced rebuild)..." -ForegroundColor Yellow
Write-Host "  Running 'go run . --generate-only'..." -ForegroundColor Gray

try {
    $output = go run . --generate-only 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Code generation completed" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Code generation failed!" -ForegroundColor Red
        Write-Host $output
        exit 1
    }
}
catch {
    Write-Host "  ✗ Code generation failed!" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit 1
}

Write-Host ""

# Step 2: Tidy dependencies
Write-Host "Step 2/4: Tidying dependencies..." -ForegroundColor Yellow
Write-Host "  Running 'go mod tidy'..." -ForegroundColor Gray

try {
    go mod tidy 2>&1 | Out-Null
    Write-Host "  ✓ Dependencies tidied" -ForegroundColor Green
}
catch {
    Write-Host "  ⚠ Warning: go mod tidy failed (continuing anyway)" -ForegroundColor Yellow
}

Write-Host ""

# Step 3: Build binary
Write-Host "Step 3/4: Building binary for $TargetOS..." -ForegroundColor Yellow

$BinaryName = ""
$BuildSuccess = $false

switch ($TargetOS.ToLower()) {
    "linux" {
        Write-Host "  Building for Linux..." -ForegroundColor Gray
        $BinaryName = "app-linux"
        $env:GOOS = "linux"
        $env:GOARCH = "amd64"
        try {
            go build -o $BinaryName .
            if ($LASTEXITCODE -eq 0) {
                $BuildSuccess = $true
                Write-Host "  ✓ Build successful: $BinaryName" -ForegroundColor Green
            }
        } catch {
            Write-Host "  ✗ Build failed!" -ForegroundColor Red
            exit 1
        }
    }
    "windows" {
        Write-Host "  Building for Windows..." -ForegroundColor Gray
        $BinaryName = "app-windows.exe"
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        try {
            go build -o $BinaryName .
            if ($LASTEXITCODE -eq 0) {
                $BuildSuccess = $true
                Write-Host "  ✓ Build successful: $BinaryName" -ForegroundColor Green
            }
        } catch {
            Write-Host "  ✗ Build failed!" -ForegroundColor Red
            exit 1
        }
    }
    "darwin" {
        Write-Host "  Building for macOS..." -ForegroundColor Gray
        $BinaryName = "app-darwin"
        $env:GOOS = "darwin"
        $env:GOARCH = "amd64"
        try {
            go build -o $BinaryName .
            if ($LASTEXITCODE -eq 0) {
                $BuildSuccess = $true
                Write-Host "  ✓ Build successful: $BinaryName" -ForegroundColor Green
            }
        } catch {
            Write-Host "  ✗ Build failed!" -ForegroundColor Red
            exit 1
        }
    }
    default {
        Write-Host "  Building for current platform..." -ForegroundColor Gray
        $BinaryName = "app.exe"
        try {
            go build -o $BinaryName .
            if ($LASTEXITCODE -eq 0) {
                $BuildSuccess = $true
                Write-Host "  ✓ Build successful: $BinaryName" -ForegroundColor Green
            }
        } catch {
            Write-Host "  ✗ Build failed!" -ForegroundColor Red
            exit 1
        }
    }
}

# Reset environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

if (-not $BuildSuccess) {
    Write-Host "  ✗ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║  Build Complete!                                   ║" -ForegroundColor Green
Write-Host "╚════════════════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""
Write-Host "Binary created: " -NoNewline
Write-Host ".\$BinaryName" -ForegroundColor Cyan
Write-Host ""
Write-Host "Usage:" -ForegroundColor Yellow
Write-Host "  .\build.ps1           - Build for current platform (Windows)"
Write-Host "  .\build.ps1 windows   - Build for Windows"
Write-Host "  .\build.ps1 linux     - Build for Linux"
Write-Host "  .\build.ps1 darwin    - Build for macOS"
Write-Host ""
