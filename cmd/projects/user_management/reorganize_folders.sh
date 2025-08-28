#!/bin/bash

# Folder Reorganization Script
# This script will reorganize the template structure for better maintainability

echo "🚀 Starting folder reorganization..."

# Base path
BASE_PATH="c:/Users/prima/SynologyDrive/golang/lokstra/cmd/projects/user_management"
OLD_TEMPLATES="$BASE_PATH/handlers/templates"
NEW_TEMPLATES="$BASE_PATH/templates"

# Create new directory structure
echo "📁 Creating new directory structure..."
mkdir -p "$NEW_TEMPLATES/layouts"
mkdir -p "$NEW_TEMPLATES/pages"
mkdir -p "$NEW_TEMPLATES/components"
mkdir -p "$NEW_TEMPLATES/assets"
mkdir -p "$NEW_TEMPLATES/assets/page-styles"

# Move layout files
echo "📄 Moving layout files..."
cp "$OLD_TEMPLATES/base_layout.html" "$NEW_TEMPLATES/layouts/base.html"
cp "$OLD_TEMPLATES/sidebar.html" "$NEW_TEMPLATES/layouts/sidebar.html"
cp "$OLD_TEMPLATES/meta_tags.html" "$NEW_TEMPLATES/layouts/meta_tags.html"

# Move page files
echo "📄 Moving page files..."
cp "$OLD_TEMPLATES/dashboard.html" "$NEW_TEMPLATES/pages/dashboard.html"
cp "$OLD_TEMPLATES/users.html" "$NEW_TEMPLATES/pages/users.html"
cp "$OLD_TEMPLATES/user-form.html" "$NEW_TEMPLATES/pages/user-form.html"
cp "$OLD_TEMPLATES/roles.html" "$NEW_TEMPLATES/pages/roles.html"
cp "$OLD_TEMPLATES/settings.html" "$NEW_TEMPLATES/pages/settings.html"

# Move component files
echo "📄 Moving component files..."
if [ -f "$OLD_TEMPLATES/templates/form.html" ]; then
    cp "$OLD_TEMPLATES/templates/form.html" "$NEW_TEMPLATES/components/forms.html"
fi
if [ -f "$OLD_TEMPLATES/templates/table.html" ]; then
    cp "$OLD_TEMPLATES/templates/table.html" "$NEW_TEMPLATES/components/tables.html"
fi
if [ -f "$OLD_TEMPLATES/templates/components.html" ]; then
    cp "$OLD_TEMPLATES/templates/components.html" "$NEW_TEMPLATES/components/common.html"
fi

# Move asset files
echo "📄 Moving asset files..."
cp "$OLD_TEMPLATES/scripts.html" "$NEW_TEMPLATES/assets/scripts.html"
cp "$OLD_TEMPLATES/styles.html" "$NEW_TEMPLATES/assets/styles.html"
if [ -d "$OLD_TEMPLATES/page-styles" ]; then
    cp -r "$OLD_TEMPLATES/page-styles/"* "$NEW_TEMPLATES/assets/page-styles/"
fi

# Copy README to new location
if [ -f "$OLD_TEMPLATES/README.md" ]; then
    cp "$OLD_TEMPLATES/README.md" "$NEW_TEMPLATES/README.md"
fi

echo "✅ Files moved successfully!"
echo "⚠️  Next steps:"
echo "   1. Update template_loader.go with new paths"
echo "   2. Test all templates work correctly"
echo "   3. Remove old folder structure after verification"

echo "🎯 New structure created at: $NEW_TEMPLATES"
