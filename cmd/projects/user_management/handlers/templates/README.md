# 📁 Template Structure Documentation

## Overview
This project uses a well-organized template structure for better maintainability, scalability, and developer experience.

## Directory Structure

```
templates/
├── layouts/          # Layout templates (base structure)
│   ├── base.html        # Main layout with header, sidebar, content
│   ├── sidebar.html     # Navigation sidebar component
│   └── meta_tags.html   # HTML head meta tags
├── pages/            # Page-specific templates
│   ├── dashboard.html   # Dashboard page content
│   ├── users.html       # User list page
│   ├── user-form.html   # User create/edit form
│   ├── roles.html       # Roles management page
│   └── settings.html    # Settings page
├── components/       # Reusable UI components
│   ├── forms.html       # Form components (inputs, buttons, etc.)
│   ├── tables.html      # Table components (data tables, pagination)
│   └── common.html      # Common UI elements (alerts, modals, etc.)
└── assets/           # Static assets and styling
    ├── scripts.html     # JavaScript includes and inline scripts
    ├── styles.html      # CSS includes and inline styles
    └── page-styles/     # Page-specific CSS files
        ├── users.css    # Styles specific to users page
        └── user-form.css # Styles specific to user form
```

## Template Categories

### 1. **Layouts** (`layouts/`)
Base structure templates that define the overall page layout.

- **base.html**: Main HTML structure with head, body, sidebar, and content areas
- **sidebar.html**: Navigation menu and sidebar component
- **meta_tags.html**: HTML meta tags for SEO and responsive design

### 2. **Pages** (`pages/`)
Content templates for specific pages or routes.

- **dashboard.html**: Main dashboard with statistics and overview
- **users.html**: User listing with table and pagination
- **user-form.html**: User creation and editing form
- **roles.html**: Role management interface
- **settings.html**: Application settings page

### 3. **Components** (`components/`)
Reusable UI components that can be included in multiple pages.

- **forms.html**: Form elements (input fields, buttons, validation)
- **tables.html**: Data table components with sorting and pagination
- **common.html**: Common UI elements (alerts, modals, breadcrumbs)

### 4. **Assets** (`assets/`)
Static assets, scripts, and styling files.

- **scripts.html**: JavaScript libraries and application scripts
- **styles.html**: CSS frameworks and custom styles
- **page-styles/**: Page-specific CSS files for fine-tuned styling

## Usage Examples

### Loading a Page Template
```go
// Render a complete page
content := renderPageContent("users", userData)
```

### Using Components
```go
// Render a reusable component
formComponent := renderComponent("forms", formData)
```

### Including Assets
```go
// Assets are automatically included in all pages
// Page-specific styles are loaded based on currentPage
```

## Benefits

### 1. **Clear Separation of Concerns**
- Layout logic separated from content
- Reusable components for consistency
- Assets organized by type

### 2. **Easy Maintenance**
- Find templates quickly by category
- Modify layouts without touching pages
- Update components once, affect all users

### 3. **Scalability**
- Add new pages without duplicating layout code
- Create new components for common patterns
- Organize assets logically

### 4. **Developer Experience**
- Intuitive folder structure
- Self-documenting organization
- Faster development and debugging

## File Naming Conventions

### Templates
- Use kebab-case: `user-form.html`, `meta-tags.html`
- Descriptive names: `dashboard.html`, `settings.html`
- Component names should be plural: `forms.html`, `tables.html`

### CSS Files
- Match the page name: `users.css` for `users.html`
- Use kebab-case: `user-form.css`

### Template Variables
- Use camelCase in Go: `userData`, `pageTitle`
- Use kebab-case in HTML: `data-user-id`, `class="user-form"`

## Adding New Templates

### New Page
1. Create `templates/pages/new-page.html`
2. Add CSS file: `templates/assets/page-styles/new-page.css`
3. Update `template_loader.go` to include the new page
4. Add route handler in appropriate handler file

### New Component
1. Create `templates/components/new-component.html`
2. Update `template_loader.go` to load the component
3. Use `renderComponent("new-component", data)` in pages

### New Layout
1. Create `templates/layouts/new-layout.html`
2. Update `template_loader.go` if needed
3. Reference in base layout or use directly

## Migration Notes

This structure replaces the old nested `handlers/templates/templates/` structure with a cleaner, more intuitive organization. All functionality remains the same, but maintenance and development are significantly improved.

## Best Practices

1. **Keep layouts minimal** - Focus on structure, not content
2. **Make components reusable** - Avoid page-specific logic in components  
3. **Use consistent naming** - Follow the established conventions
4. **Document complex templates** - Add comments for complex logic
5. **Test after changes** - Verify all pages still render correctly
