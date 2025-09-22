# Admin Dashboard App

A professional admin dashboard application built with Lokstra and HTMX.

## Features

- **Dashboard**: Real-time statistics and activity feed
- **User Management**: View and manage system users
- **Settings**: System configuration panel
- **Analytics**: Traffic and performance metrics
- **API Endpoints**: RESTful APIs for dynamic content

## Theme

Professional blue color scheme with modern UI components.

## Structure

```
app1_admin/
├── admin_app.go          # Main application setup and routing
├── layouts/
│   └── admin.html        # Base layout template
├── pages/
│   ├── dashboard.html    # Dashboard page
│   ├── users.html        # User management page
│   ├── settings.html     # Settings page
│   └── analytics.html    # Analytics page
├── static/
│   ├── admin.css         # Stylesheet
│   └── admin.js          # JavaScript functionality
└── README.md            # This file
```

## Usage

Import and setup the admin app in your main application:

```go
package main

import (
    "github.com/primadi/lokstra"
    "path/to/your/project/app1_admin"
)

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    app := lokstra.NewApp(regCtx, "multi-app", ":8080")
    
    // Setup admin app with /admin prefix
    admin.Setup(app, "/admin")
    
    app.Start(true)
}
```

## Routes

- `GET /admin/` - Dashboard
- `GET /admin/users` - User management
- `GET /admin/settings` - System settings
- `GET /admin/analytics` - Analytics
- `GET /admin/api/stats` - Live statistics API
- `POST /admin/api/users` - Create user API

## Dependencies

- Lokstra framework
- HTMX for dynamic interactions
- Embedded file system for deployment

## License

Same as the parent project.