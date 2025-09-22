package blog

import (
	"embed"
	"fmt"
	"strings"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/htmx_fsmanager"
	"github.com/primadi/lokstra/core/router"
)

//go:embed layouts/* pages/* static/*
var AppFiles embed.FS

// Setup configures the blog platform application with the given prefix
func Setup(app *lokstra.App, prefix string) {
	blogGroup := app.Group(prefix)

	// Setup script injection for HTMX with default scripts that handle LS-Title
	scriptInjection := htmx_fsmanager.NewDefaultScriptInjection(true)

	// Setup HTMX manager with embedded files and script injection
	blogGroup.SetHTMXLayoutScriptInjection(scriptInjection).
		AddHtmxLayouts(AppFiles, "layouts").
		AddHtmxPages(AppFiles, "pages").
		AddHtmxStatics(AppFiles, "static")

	setupRoutes(blogGroup)
}

func setupRoutes(group router.Router) {
	// Home page
	group.GET("/", func(ctx *lokstra.Context) error {
		// Hero section data
		hero := map[string]interface{}{
			"title":    "Welcome to TechBlog",
			"subtitle": "Sharing knowledge, inspiring growth",
		}

		// Stats data
		stats := map[string]interface{}{
			"total_articles": 127,
			"total_readers":  2485,
			"total_comments": 892,
		}

		// Featured articles data
		featuredArticles := []map[string]interface{}{
			{
				"slug":     "getting-started-htmx-go",
				"title":    "Getting Started with HTMX and Go",
				"excerpt":  "Learn how to build modern web applications using HTMX with a Go backend.",
				"category": "Web Development",
				"image":    "/static/images/article1.jpg",
				"author": map[string]interface{}{
					"name":   "Alex Thompson",
					"avatar": "👨‍💻",
				},
				"published_date": "2025-09-20",
				"read_time":      5,
				"views":          1234,
				"tags":           []string{"HTMX", "Go", "Web Development"},
			},
			{
				"slug":     "css-grid-responsive-layouts",
				"title":    "Building Responsive Layouts with CSS Grid",
				"excerpt":  "Master CSS Grid to create beautiful, responsive layouts for modern web applications.",
				"category": "CSS & Design",
				"image":    "/static/images/article2.jpg",
				"author": map[string]interface{}{
					"name":   "Sarah Chen",
					"avatar": "👩‍🎨",
				},
				"published_date": "2025-09-18",
				"read_time":      8,
				"views":          987,
				"tags":           []string{"CSS", "Grid", "Responsive Design"},
			},
			{
				"slug":     "javascript-modern-patterns",
				"title":    "Modern JavaScript Patterns",
				"excerpt":  "Explore the latest JavaScript patterns and techniques for building scalable applications.",
				"category": "JavaScript",
				"image":    "/static/images/article3.jpg",
				"author": map[string]interface{}{
					"name":   "Mike Rodriguez",
					"avatar": "⚡",
				},
				"published_date": "2025-09-15",
				"read_time":      6,
				"views":          756,
				"tags":           []string{"JavaScript", "Patterns", "Modern Development"},
			},
		}

		// Popular categories data
		popularCategories := []map[string]interface{}{
			{
				"slug":          "web-development",
				"name":          "Web Development",
				"description":   "Frontend and backend development tutorials",
				"icon":          "🌐",
				"article_count": 45,
			},
			{
				"slug":          "javascript",
				"name":          "JavaScript",
				"description":   "Modern JavaScript techniques and frameworks",
				"icon":          "⚡",
				"article_count": 32,
			},
			{
				"slug":          "css-design",
				"name":          "CSS & Design",
				"description":   "Styling and user interface design",
				"icon":          "🎨",
				"article_count": 28,
			},
			{
				"slug":          "go-programming",
				"name":          "Go Programming",
				"description":   "Go language tutorials and best practices",
				"icon":          "🚀",
				"article_count": 22,
			},
		}

		data := map[string]interface{}{
			"hero":               hero,
			"stats":              stats,
			"featured_articles":  featuredArticles,
			"popular_categories": popularCategories,
		}

		return ctx.HTMXFSPage("home", data, "Home - Tech Blog", "Welcome to our tech blog platform.")
	})

	// Articles page
	group.GET("/articles", func(ctx *lokstra.Context) error {
		articles := []map[string]interface{}{
			{
				"id":       1,
				"title":    "Getting Started with HTMX and Go",
				"excerpt":  "Learn how to build modern web applications using HTMX with a Go backend.",
				"author":   "Tech Blogger",
				"date":     "2025-09-20",
				"readTime": "5 min read",
				"tags":     []string{"HTMX", "Go", "Web Development"},
			},
			{
				"id":       2,
				"title":    "Building Responsive Layouts with CSS Grid",
				"excerpt":  "Master CSS Grid to create beautiful, responsive layouts for modern web applications.",
				"author":   "Design Expert",
				"date":     "2025-09-18",
				"readTime": "8 min read",
				"tags":     []string{"CSS", "Grid", "Responsive Design"},
			},
			{
				"id":       3,
				"title":    "Modern JavaScript Patterns",
				"excerpt":  "Explore the latest JavaScript patterns and techniques for building scalable applications.",
				"author":   "JS Developer",
				"date":     "2025-09-15",
				"readTime": "6 min read",
				"tags":     []string{"JavaScript", "Patterns", "Modern Development"},
			},
		}

		data := map[string]interface{}{"articles": articles}
		return ctx.HTMXFSPage("articles", data, "Articles - Tech Blog", "Browse all articles on our tech blog platform.")
	})

	// Categories page
	group.GET("/categories", func(ctx *lokstra.Context) error {
		categories := []map[string]interface{}{
			{"name": "Web Development", "count": 45, "description": "Frontend and backend development tutorials"},
			{"name": "JavaScript", "count": 32, "description": "Modern JavaScript techniques and frameworks"},
			{"name": "CSS & Design", "count": 28, "description": "Styling and user interface design"},
			{"name": "Go Programming", "count": 22, "description": "Go language tutorials and best practices"},
		}

		data := map[string]interface{}{"categories": categories}
		return ctx.HTMXFSPage("categories", data, "Categories - Tech Blog", "Explore article categories on our tech blog platform.")
	})

	// About page
	group.GET("/about", func(ctx *lokstra.Context) error {
		team := []map[string]interface{}{
			{"name": "Alex Thompson", "role": "Founder & Lead Developer", "bio": "Full-stack developer with 10+ years experience"},
			{"name": "Sarah Chen", "role": "Content Creator", "bio": "Technical writer and developer advocate"},
			{"name": "Mike Rodriguez", "role": "UI/UX Designer", "bio": "Passionate about creating beautiful user experiences"},
		}

		data := map[string]interface{}{
			"team":    team,
			"founded": "2023",
			"mission": "To share knowledge and help developers build better web applications",
		}
		return ctx.HTMXFSPage("about", data, "About - Tech Blog", "Learn more about our team and mission.")
	})

	// API endpoints
	group.GET("/api/search", func(ctx *lokstra.Context) error {
		query := ctx.GetQueryParam("q")
		// Simulate search results
		results := []map[string]interface{}{
			{"title": "Search result for: " + query, "excerpt": "Mock search result", "url": "/articles/1"},
		}
		return ctx.Ok(results)
	})

	// Activity feed endpoint
	group.GET("/api/activity", func(ctx *lokstra.Context) error {
		activities := []map[string]interface{}{
			{
				"type":      "new_article",
				"message":   "New article published: 'Advanced HTMX Patterns'",
				"timestamp": "2 minutes ago",
				"icon":      "📝",
			},
			{
				"type":      "comment",
				"message":   "Sarah commented on 'Getting Started with Go'",
				"timestamp": "15 minutes ago",
				"icon":      "💬",
			},
			{
				"type":      "like",
				"message":   "5 people liked 'CSS Grid Layout Guide'",
				"timestamp": "1 hour ago",
				"icon":      "❤️",
			},
		}

		var htmlBuilder strings.Builder
		htmlBuilder.WriteString(`<div class="activity-list">`)
		for _, activity := range activities {
			htmlBuilder.WriteString(fmt.Sprintf(`
				<div class="activity-item">
					<span class="activity-icon">%s</span>
					<div class="activity-content">
						<p>%s</p>
						<span class="activity-time">%s</span>
					</div>
				</div>`, activity["icon"], activity["message"], activity["timestamp"]))
		}
		htmlBuilder.WriteString(`</div>`)

		return ctx.HTML(200, htmlBuilder.String())
	})

	// Featured post endpoint
	group.GET("/api/featured", func(ctx *lokstra.Context) error {
		featured := map[string]interface{}{
			"title":    "Building Modern Web Apps with HTMX",
			"excerpt":  "Learn how to create dynamic web applications without complex JavaScript frameworks.",
			"author":   "Alex Thompson",
			"date":     "2025-09-20",
			"readTime": "8 min read",
			"image":    "/static/images/featured.jpg",
			"url":      "/app2/articles/1",
		}

		html := fmt.Sprintf(`
			<div class="featured-article">
				<div class="featured-image">
					<img src="%s" alt="%s" loading="lazy">
				</div>
				<div class="featured-content">
					<h4><a href="%s">%s</a></h4>
					<p>%s</p>
					<div class="featured-meta">
						<span>📅 %s</span>
						<span>⏱️ %s</span>
					</div>
				</div>
			</div>
		`, featured["image"], featured["title"], featured["url"], featured["title"],
			featured["excerpt"], featured["date"], featured["readTime"])

		return ctx.HTML(200, html)
	})

	// Recent posts endpoint
	group.GET("/api/recent", func(ctx *lokstra.Context) error {
		recent := []map[string]interface{}{
			{
				"title":    "CSS Grid vs Flexbox",
				"date":     "2025-09-19",
				"url":      "/app2/articles/2",
				"comments": 12,
			},
			{
				"title":    "JavaScript ES2024 Features",
				"date":     "2025-09-18",
				"url":      "/app2/articles/3",
				"comments": 8,
			},
			{
				"title":    "Go Best Practices",
				"date":     "2025-09-17",
				"url":      "/app2/articles/4",
				"comments": 15,
			},
		}

		var htmlBuilder strings.Builder
		htmlBuilder.WriteString(`<div class="recent-list">`)
		for _, post := range recent {
			htmlBuilder.WriteString(fmt.Sprintf(`
				<div class="recent-item">
					<h5><a href="%s">%s</a></h5>
					<div class="recent-meta">
						<span>📅 %s</span>
						<span>💬 %d comments</span>
					</div>
				</div>`, post["url"], post["title"], post["date"], post["comments"]))
		}
		htmlBuilder.WriteString(`</div>`)

		return ctx.HTML(200, htmlBuilder.String())
	})

	// Categories endpoint
	group.GET("/api/categories", func(ctx *lokstra.Context) error {
		categories := []map[string]interface{}{
			{"name": "Web Development", "count": 45, "slug": "web-dev"},
			{"name": "JavaScript", "count": 32, "slug": "javascript"},
			{"name": "CSS & Design", "count": 28, "slug": "css-design"},
			{"name": "Go Programming", "count": 22, "slug": "go"},
			{"name": "DevOps", "count": 18, "slug": "devops"},
		}

		var htmlBuilder strings.Builder
		htmlBuilder.WriteString(`<div class="categories-widget">`)
		for _, category := range categories {
			htmlBuilder.WriteString(fmt.Sprintf(`
				<div class="category-item">
					<a href="/app2/category/%s">%s</a>
					<span class="category-count">(%d)</span>
				</div>`, category["slug"], category["name"], category["count"]))
		}
		htmlBuilder.WriteString(`</div>`)

		return ctx.HTML(200, htmlBuilder.String())
	})

	// Tags endpoint
	group.GET("/api/tags", func(ctx *lokstra.Context) error {
		tags := []map[string]interface{}{
			{"name": "HTMX", "count": 15, "slug": "htmx"},
			{"name": "React", "count": 28, "slug": "react"},
			{"name": "CSS", "count": 22, "slug": "css"},
			{"name": "Performance", "count": 18, "slug": "performance"},
			{"name": "Tutorial", "count": 35, "slug": "tutorial"},
			{"name": "Tips", "count": 20, "slug": "tips"},
			{"name": "Backend", "count": 25, "slug": "backend"},
			{"name": "Frontend", "count": 30, "slug": "frontend"},
		}

		var htmlBuilder strings.Builder
		htmlBuilder.WriteString(`<div class="tags-widget">`)
		for _, tag := range tags {
			fontSize := "1em"
			count := tag["count"].(int)
			if count > 25 {
				fontSize = "1.2em"
			} else if count > 15 {
				fontSize = "1.1em"
			}

			htmlBuilder.WriteString(fmt.Sprintf(`
				<a href="/app2/tag/%s" class="tag-item" style="font-size: %s">
					%s
				</a>`, tag["slug"], fontSize, tag["name"]))
		}
		htmlBuilder.WriteString(`</div>`)

		return ctx.HTML(200, htmlBuilder.String())
	})

	// Newsletter subscription endpoint
	group.POST("/api/newsletter/subscribe", func(ctx *lokstra.Context) error {
		email := ctx.Request.FormValue("email")
		response := fmt.Sprintf(`
			<div class="newsletter-success">
				<span class="success-icon">✅</span>
				<p>Thanks for subscribing! Welcome to our newsletter, %s</p>
			</div>
		`, email)
		return ctx.HTML(200, response)
	})

	// Keep the old newsletter endpoint for backwards compatibility
	group.POST("/api/newsletter", func(ctx *lokstra.Context) error {
		email := ctx.Request.FormValue("email")
		response := map[string]interface{}{
			"success": true,
			"message": "Thanks for subscribing! Welcome to our newsletter, " + email,
		}
		return ctx.Ok(response)
	})
}
