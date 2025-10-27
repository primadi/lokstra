# Script to rename README.md to index.md in all subdirectories
# This fixes Jekyll routing - Jekyll looks for index.md, not README.md

$docsPath = "c:\Users\prima\SynologyDrive\golang\lokstra-dev2\docs"

# Get all README.md files in subdirectories (not root)
$readmeFiles = Get-ChildItem -Path $docsPath -Filter "README.md" -Recurse | 
    Where-Object { 
        $_.DirectoryName -ne $docsPath -and  # Exclude root README.md
        $_.FullName -notlike "*\_layouts\*" -and
        $_.FullName -notlike "*\schema\*"
    }

$renamed = 0
$skipped = 0

foreach ($file in $readmeFiles) {
    $newPath = Join-Path $file.DirectoryName "index.md"
    
    if (Test-Path $newPath) {
        Write-Host "SKIP: $($file.DirectoryName) (index.md already exists)" -ForegroundColor Yellow
        $skipped++
    } else {
        Move-Item -Path $file.FullName -Destination $newPath
        Write-Host "✓ Renamed: $($file.FullName) → index.md" -ForegroundColor Green
        $renamed++
    }
}

Write-Host "`n=== Summary ===" -ForegroundColor Magenta
Write-Host "Renamed: $renamed files" -ForegroundColor Green
Write-Host "Skipped: $skipped files" -ForegroundColor Yellow
