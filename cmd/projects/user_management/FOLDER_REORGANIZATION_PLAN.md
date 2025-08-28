# 📁 Folder Reorganization Plan

## Current Structure Issues
- Duplicate `templates/` folders: `handlers/templates/` and `handlers/templates/templates/`
- Mixed concerns: assets mixed with page templates
- Hard to navigate and maintain

## Proposed New Structure

```
cmd/projects/user_management/
├── handlers/
│   ├── ui_handlers.go
│   ├── user_handler.go
│   ├── dto.go
│   └── template_loader.go       # Updated to point to new paths
├── templates/
│   ├── layouts/
│   │   ├── base.html           # From: handlers/templates/base_layout.html
│   │   ├── sidebar.html        # From: handlers/templates/sidebar.html
│   │   └── meta_tags.html      # From: handlers/templates/meta_tags.html
│   ├── pages/
│   │   ├── dashboard.html      # From: handlers/templates/dashboard.html
│   │   ├── users.html          # From: handlers/templates/users.html
│   │   ├── user-form.html      # From: handlers/templates/user-form.html
│   │   ├── roles.html          # From: handlers/templates/roles.html
│   │   └── settings.html       # From: handlers/templates/settings.html
│   ├── components/
│   │   ├── forms.html          # From: handlers/templates/templates/form.html
│   │   ├── tables.html         # From: handlers/templates/templates/table.html
│   │   └── common.html         # From: handlers/templates/templates/components.html
│   └── assets/
│       ├── scripts.html        # From: handlers/templates/scripts.html
│       ├── styles.html         # From: handlers/templates/styles.html
│       └── page-styles/        # From: handlers/templates/page-styles/
└── static/                     # New: for static files (CSS, JS, images)
    ├── css/
    ├── js/
    └── images/
```

## Benefits

### 1. **Clear Separation of Concerns**
- `layouts/`: Base layouts and common structure
- `pages/`: Individual page templates
- `components/`: Reusable components
- `assets/`: CSS, JS, and styling assets

### 2. **Better Maintainability**
- Easy to find specific templates
- Logical grouping by function
- No duplicate folder names

### 3. **Scalability**
- Easy to add new pages in `pages/`
- Reusable components in `components/`
- Static assets properly separated

### 4. **Developer Experience**
- Intuitive folder structure
- Faster navigation
- Clear naming conventions

## Migration Steps

1. **Create new structure**
2. **Move files to appropriate locations**
3. **Update template_loader.go paths**
4. **Update import statements in handlers**
5. **Test all templates work correctly**
6. **Remove old duplicate folders**

## Template Naming Convention

### Layouts
- `base.html` - Main layout with header, sidebar, content area
- `sidebar.html` - Navigation sidebar component
- `meta_tags.html` - HTML head meta tags

### Pages
- `dashboard.html` - Dashboard page
- `users.html` - User list page
- `user-form.html` - User create/edit form
- `roles.html` - Roles management page
- `settings.html` - Settings page

### Components
- `forms.html` - Reusable form components
- `tables.html` - Reusable table components
- `common.html` - Common UI components

### Assets
- `scripts.html` - JavaScript includes
- `styles.html` - CSS includes
- `page-styles/` - Page-specific styling

## File Updates Required

1. **template_loader.go**: Update all template paths
2. **ui_handlers.go**: Update template references
3. **user_handler.go**: Update template references
