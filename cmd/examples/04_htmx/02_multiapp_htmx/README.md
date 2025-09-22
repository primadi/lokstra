# Multi-App HTMX Demo

This example demonstrates how to create multiple independent applications that can run together using the Lokstra framework with HTMX integration.

## 🎯 Key Features

- **Modular Architecture**: Each app is completely self-contained in its own folder
- **Automatic Static Path Injection**: No conflicts between apps' static files
- **Independent Deployment**: Copy any app folder to reuse in other projects
- **Theme Isolation**: Each app has its own distinct visual theme
- **HTMX Integration**: Dynamic content loading without page refreshes

## 📁 Project Structure

```
02_multiapp_htmx/
├── main.go                    # Main application entry point
├── go.mod                     # Module definition with local paths
├── README.md                  # This file
├── app1_admin/               # Admin Dashboard App
│   ├── admin_app.go          # App setup and routing
│   ├── layouts/
│   │   └── admin.html        # Admin layout template
│   ├── pages/
│   │   ├── dashboard.html    # Dashboard page
│   │   ├── users.html        # User management
│   │   ├── settings.html     # Settings page
│   │   └── analytics.html    # Analytics page
│   ├── static/
│   │   ├── admin.css         # Blue theme stylesheet
│   │   └── admin.js          # Admin JavaScript
│   └── README.md             # Admin app documentation
└── app2_blog/                # Blog Platform App
    ├── blog_app.go           # App setup and routing
    ├── layouts/
    │   └── blog.html         # Blog layout template
    ├── pages/
    │   ├── home.html         # Home page
    │   ├── articles.html     # Articles listing
    │   ├── categories.html   # Categories page
    │   └── about.html        # About page
    ├── static/
    │   ├── blog.css          # Green theme stylesheet
    │   └── blog.js           # Blog JavaScript
    └── README.md             # Blog app documentation
```

## 🚀 Running the Demo

1. **Navigate to the project directory:**
   ```bash
   cd cmd/examples/04_htmx/02_multiapp_htmx
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the application:**
   ```bash
   go run .
   ```

4. **Open your browser and visit:**
   - http://localhost:8080 - Landing page with app showcase
   - http://localhost:8080/app1 - Admin Dashboard (Blue theme)
   - http://localhost:8080/app2 - Blog Platform (Green theme)

## 🎨 Applications

### App1: Admin Dashboard (/app1)
- **Theme**: Professional blue color scheme
- **Features**: Dashboard, user management, settings, analytics
- **Purpose**: Demonstrates admin panel patterns

### App2: Blog Platform (/app2)
- **Theme**: Creative green color scheme
- **Features**: Articles, categories, about page, newsletter
- **Purpose**: Demonstrates content management patterns

## 📦 Reusing Apps in Other Projects

Each app folder is completely self-contained and can be copied to other projects:

### Option 1: Copy and Import

```bash
# Copy app folder to your project
cp -r app1_admin /path/to/your/project/

# In your main.go
import "yourproject/app1_admin"

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    app := lokstra.NewApp(regCtx, "my-app", ":8080")
    
    // Setup admin with custom prefix
    admin.Setup(app, "/admin")
    
    app.Start(true)
}
```

### Option 2: Go Module

```bash
# Create a separate module for the app
cd app1_admin
go mod init github.com/yourname/admin-app

# Use in other projects
go get github.com/yourname/admin-app
```

## 🔧 Customization

### Changing App Prefix

```go
// Instead of /app1, use /admin
admin.Setup(app, "/admin")

// Instead of /app2, use /blog  
blog.Setup(app, "/blog")
```

### Adding New Routes

Edit the `setupRoutes` function in each app's main file:

```go
// In app1_admin/admin_app.go
func setupRoutes(group router.Router) {
    // Add your new routes here
    group.GET("/new-page", handler)
}
```

### Modifying Themes

- Edit CSS files in the `static/` folder
- Update layout templates in the `layouts/` folder
- Add new pages in the `pages/` folder

## 📚 Learning Objectives

1. **Multi-App Architecture**: How to structure multiple applications
2. **Static Path Injection**: Automatic handling of static file routing
3. **App Isolation**: Preventing conflicts between applications
4. **HTMX Integration**: Dynamic content without complex JavaScript
5. **Embedded File Systems**: Production-ready deployment patterns

## 🛠️ Technical Implementation

### Automatic Static Path Injection

The framework automatically:
- Detects static files in each app
- Creates appropriate routes with prefixes
- Injects correct static paths in templates
- Prevents conflicts between apps

### Embedded File Systems

Each app uses Go's `embed` package to include all assets:
- Templates are embedded for production deployment
- Static files are embedded and served efficiently
- No external file dependencies in production

### HTMX Integration

- Server-side rendering with dynamic updates
- Form handling with real-time feedback
- Navigation without page refreshes
- API endpoints for dynamic content

## 📖 Documentation

- [Admin App Details](app1_admin/README.md)
- [Blog App Details](app2_blog/README.md)
- [Lokstra HTMX Integration](https://github.com/primadi/lokstra/blob/main/docs/htmx-integration.md)

## 🤝 Contributing

This example is part of the Lokstra framework documentation. For improvements or bug fixes, please contribute to the main repository.

## 📄 License

Same as the Lokstra framework license.