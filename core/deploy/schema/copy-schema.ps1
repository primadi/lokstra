#!/usr/bin/env pwsh
# copy-schema.ps1
# Copies lokstra.schema.json to docs/schema/ for GitHub Pages publishing

$ErrorActionPreference = "Stop"

# Get script directory (core/deploy/schema)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Calculate paths
$sourceFile = Join-Path $scriptDir "lokstra.schema.json"
$targetDir = Join-Path $scriptDir "..\..\..\docs\schema"
$targetFile = Join-Path $targetDir "lokstra.schema.json"

# Normalize paths
$sourceFile = [System.IO.Path]::GetFullPath($sourceFile)
$targetDir = [System.IO.Path]::GetFullPath($targetDir)
$targetFile = [System.IO.Path]::GetFullPath($targetFile)

Write-Host "📋 Copying schema file..." -ForegroundColor Cyan
Write-Host "   Source: $sourceFile" -ForegroundColor Gray
Write-Host "   Target: $targetFile" -ForegroundColor Gray

# Check if source exists
if (-not (Test-Path $sourceFile)) {
    Write-Host "❌ ERROR: Source file not found: $sourceFile" -ForegroundColor Red
    exit 1
}

# Create target directory if not exists
if (-not (Test-Path $targetDir)) {
    Write-Host "📁 Creating directory: $targetDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
}

# Copy file
try {
    Copy-Item -Path $sourceFile -Destination $targetFile -Force
    Write-Host "✅ Schema copied successfully!" -ForegroundColor Green
    Write-Host "   📍 Location: $targetFile" -ForegroundColor Gray
    
    # Show file size
    $fileSize = (Get-Item $targetFile).Length
    $fileSizeKB = [Math]::Round($fileSize / 1KB, 2)
    Write-Host "   📊 Size: $fileSizeKB KB" -ForegroundColor Gray
    
} catch {
    Write-Host "❌ ERROR: Failed to copy file" -ForegroundColor Red
    Write-Host "   $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🎉 Done! Schema ready for GitHub Pages deployment." -ForegroundColor Green
Write-Host "   Run 'git add docs/schema/lokstra.schema.json' to stage changes." -ForegroundColor Gray
