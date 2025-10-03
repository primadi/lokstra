# Lokstra Router Integration Test Script
# This script helps you run different deployment modes easily

param(
    [Parameter(Position=0)]
    [string]$DeploymentId = "",
    [string]$ServerName = "all"
)

# Function to show usage
function Show-Usage {
    Write-Host ""
    Write-Host "🚀 Lokstra Router Integration Test Runner" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Yellow
    Write-Host "  .\run.ps1 [deployment-id] [-ServerName server-name]" -ForegroundColor White
    Write-Host ""
    Write-Host "Available Deployment Types:" -ForegroundColor Cyan
    Write-Host "  monolith-single-port   - All routers in one server on one port (:8081)" -ForegroundColor White
    Write-Host "  monolith-multi-port    - All routers in one server on different ports (:8081, :8082)" -ForegroundColor White
    Write-Host "  microservice           - Each router on its own server (:8082, :8083)" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\run.ps1 monolith-single-port" -ForegroundColor White
    Write-Host "  .\run.ps1 monolith-multi-port" -ForegroundColor White
    Write-Host "  .\run.ps1 microservice" -ForegroundColor White
    Write-Host "  .\run.ps1 microservice -ServerName product-service" -ForegroundColor White
    Write-Host ""
}

# Show usage if no deployment ID provided
if ($DeploymentId -eq "") {
    Show-Usage
    $DeploymentId = Read-Host "Enter deployment ID"
    if ($DeploymentId -eq "") {
        Write-Host "❌ No deployment ID provided. Exiting." -ForegroundColor Red
        exit 1
    }
}

# Validate deployment ID
$ValidDeployments = @("monolith-single-port", "monolith-multi-port", "microservice")
if ($DeploymentId -notin $ValidDeployments) {
    Write-Host "❌ Invalid deployment ID: $DeploymentId" -ForegroundColor Red
    Write-Host "Valid options: $($ValidDeployments -join ', ')" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "🎯 Starting deployment: $DeploymentId" -ForegroundColor Green
Write-Host "🖥️  Server filter: $ServerName" -ForegroundColor Green
Write-Host ""

# Set environment variables
$env:DEPLOYMENT_ID = $DeploymentId
$env:SERVER_NAME = $ServerName

# Change to the correct directory
Set-Location $PSScriptRoot

# Build and run
Write-Host "🔨 Building application..." -ForegroundColor Yellow
go build -o router-integration.exe .

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Build successful!" -ForegroundColor Green
Write-Host ""
Write-Host "🚀 Starting server(s)..." -ForegroundColor Yellow

# Run the application
.\router-integration.exe