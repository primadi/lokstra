# Tech Blog App

A modern blog platform application built with Lokstra and HTMX.

## Features

- **Home Page**: Featured articles and blog statistics
- **Articles**: Browse and search articles
- **Categories**: Organized content categories
- **About**: Team information and mission
- **Newsletter**: Email subscription functionality
- **Search**: Real-time article search

## Theme

Creative green color scheme with modern, responsive design.

## Structure

```
app2_blog/
├── blog_app.go           # Main application setup and routing
├── layouts/
│   └── blog.html         # Base layout template
├── pages/
│   ├── home.html         # Home page
│   ├── articles.html     # Articles listing page
│   ├── categories.html   # Categories page
│   └── about.html        # About page
├── static/
│   ├── blog.css          # Stylesheet
│   └── blog.js           # JavaScript functionality
└── README.md            # This file
```

## Usage

Import and setup the blog app in your main application:

```go
package main

import (
    "github.com/primadi/lokstra"
    "path/to/your/project/app2_blog"
)

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    app := lokstra.NewApp(regCtx, "multi-app", ":8080")
    
    // Setup blog app with /blog prefix
    blog.Setup(app, "/blog")
    
    app.Start(true)
}
```

## Routes

- `GET /blog/` - Home page with featured articles
- `GET /blog/articles` - Articles listing
- `GET /blog/categories` - Categories page
- `GET /blog/about` - About page
- `GET /blog/api/search` - Search API
- `POST /blog/api/newsletter` - Newsletter subscription API

## Features

### Frontend
- Responsive design with mobile support
- Dark/light theme toggle
- Search overlay with keyboard shortcuts
- Reading progress indicator
- Social sharing functionality
- Infinite scroll for articles
- Form validation

### Backend
- RESTful API endpoints
- Real-time search
- Newsletter subscription handling
- Content categorization

## Dependencies

- Lokstra framework
- HTMX for dynamic interactions
- Embedded file system for deployment

## License

Same as the parent project.