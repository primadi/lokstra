#!/usr/bin/env pwsh
# Demo script untuk testing scalable e-commerce deployment

Write-Host "🏪 E-Commerce Scalability Demo" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

Write-Host "`n📋 Available Demo Scenarios:" -ForegroundColor Yellow
Write-Host "1. Monolith Deployment (All services in single server)"
Write-Host "2. Product Microservice (Product catalog service only)" 
Write-Host "3. Order Microservice (Order processing service only)"
Write-Host "4. User Microservice (User management service only)"
Write-Host "5. Payment Microservice (Payment processing service only)"
Write-Host "6. Analytics Microservice (Analytics & reporting service only)"

$choice = Read-Host "`nSelect deployment scenario (1-6)"

switch ($choice) {
    "1" {
        Write-Host "`n🚀 Starting MONOLITH deployment..." -ForegroundColor Green
        Write-Host "Single server with ALL services and APIs" -ForegroundColor Gray
        $env:DEPLOYMENT_TYPE = "monolith"
        Remove-Variable SERVER_NAME -ErrorAction SilentlyContinue
        go run main.go
    }
    "2" {
        Write-Host "`n🚀 Starting PRODUCT microservice..." -ForegroundColor Green
        Write-Host "Independent product catalog service" -ForegroundColor Gray
        $env:DEPLOYMENT_TYPE = "microservices"
        $env:SERVER_NAME = "product-service"
        go run main.go
    }
    "3" {
        Write-Host "`n🚀 Starting ORDER microservice..." -ForegroundColor Green
        Write-Host "Independent order processing service" -ForegroundColor Gray
        $env:DEPLOYMENT_TYPE = "microservices"
        $env:SERVER_NAME = "order-service"
        go run main.go
    }
    "4" {
        Write-Host "`n🚀 Starting USER microservice..." -ForegroundColor Green
        Write-Host "Independent user management service" -ForegroundColor Gray
        $env:DEPLOYMENT_TYPE = "microservices"
        $env:SERVER_NAME = "user-service"
        go run main.go
    }
    "5" {
        Write-Host "`n🚀 Starting PAYMENT microservice..." -ForegroundColor Green
        Write-Host "Independent payment processing service" -ForegroundColor Gray
        $env:DEPLOYMENT_TYPE = "microservices"
        $env:SERVER_NAME = "payment-service"
        go run main.go
    }
    "6" {
        Write-Host "`n🚀 Starting ANALYTICS microservice..." -ForegroundColor Green
        Write-Host "Independent analytics & reporting service" -ForegroundColor Gray
        $env:DEPLOYMENT_TYPE = "microservices"
        $env:SERVER_NAME = "analytics-service"
        go run main.go
    }
    default {
        Write-Host "❌ Invalid choice. Please select 1-6." -ForegroundColor Red
        exit 1
    }
}