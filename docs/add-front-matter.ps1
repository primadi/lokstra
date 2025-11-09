# Script to add front matter to all markdown files in docs folder
# This adds "layout: docs" and title extracted from first H1

$docsPath = "c:\Users\prima\SynologyDrive\golang\lokstra-dev2\docs"

# Get all markdown files, excluding index.md (uses default layout)
$mdFiles = Get-ChildItem -Path $docsPath -Filter "*.md" -Recurse | 
    Where-Object { 
        $_.Name -ne "index.md" -and 
        $_.FullName -notlike "*\_layouts\*" -and
        $_.FullName -notlike "*\schema\*"
    }

$updated = 0
$skipped = 0
$errors = 0

foreach ($file in $mdFiles) {
    try {
        $content = Get-Content -Path $file.FullName -Raw
        
        # Skip if already has front matter
        if ($content -match "^---\s*\nlayout:") {
            Write-Host "SKIP: $($file.FullName) (already has front matter)" -ForegroundColor Yellow
            $skipped++
            continue
        }
        
        # Extract title from first H1
        if ($content -match "^#\s+(.+)$" -or $content -match "\n#\s+(.+)$") {
            $title = $matches[1].Trim()
            
            # Create front matter
            $frontMatter = @"
---
layout: docs
title: $title
---

"@
            
            # Add front matter to beginning of file
            $newContent = $frontMatter + $content
            
            # Write back to file
            Set-Content -Path $file.FullName -Value $newContent -NoNewline
            
            Write-Host "✓ Updated: $($file.FullName)" -ForegroundColor Green
            $updated++
        } else {
            Write-Host "WARN: $($file.FullName) (no H1 found)" -ForegroundColor Cyan
            $skipped++
        }
    } catch {
        Write-Host "ERROR: $($file.FullName) - $($_.Exception.Message)" -ForegroundColor Red
        $errors++
    }
}

Write-Host "`n=== Summary ===" -ForegroundColor Magenta
Write-Host "Updated: $updated files" -ForegroundColor Green
Write-Host "Skipped: $skipped files" -ForegroundColor Yellow
Write-Host "Errors:  $errors files" -ForegroundColor Red
