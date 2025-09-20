package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/primadi/lokstra"
)

func createApiRoutes(app *lokstra.App) {
	// ===== Dynamic Content API =====

	// API routes for dynamic content loading
	api := app.Group("/api")

	// Dynamic content for home page
	api.GET("/featured-content", func(ctx *lokstra.Context) error {
		htmlContent := `
<div class="featured-items">
    <h4>Featured Content</h4>
    <div class="featured-grid">
        <div class="featured-item-card">
            <h5>Getting Started Guide</h5>
            <p>Learn the basics of Lokstra</p>
            <a href="/docs/getting-started" class="btn btn-sm btn-outline">Learn More</a>
        </div>
        <div class="featured-item-card">
            <h5>HTMX Tutorial</h5>
            <p>Master HTMX with practical examples</p>
            <a href="/docs/htmx" class="btn btn-sm btn-outline">Learn More</a>
        </div>
        <div class="featured-item-card">
            <h5>Best Practices</h5>
            <p>Tips for building great apps</p>
            <a href="/docs/best-practices" class="btn btn-sm btn-outline">Learn More</a>
        </div>
    </div>
</div>

<style>
.featured-items {
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 1.5rem;
}

.featured-items h4 {
    margin: 0 0 1rem 0;
    color: #007bff;
}

.featured-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1rem;
}

.featured-item-card {
    background: #f8f9fa;
    border: 1px solid #dee2e6;
    border-radius: 6px;
    padding: 1rem;
}

.featured-item-card h5 {
    margin: 0 0 0.5rem 0;
    color: #212529;
}

.featured-item-card p {
    margin: 0 0 1rem 0;
    color: #6c757d;
    font-size: 0.875rem;
}
</style>
`
		return ctx.HTML(http.StatusOK, htmlContent)
	})

	// Dynamic stats
	api.GET("/stats", func(ctx *lokstra.Context) error {
		htmlContent := fmt.Sprintf(`
<div class="stats-live">
    <div class="stats-cards">
        <div class="stat-card">
            <div class="stat-icon">👥</div>
            <div class="stat-info">
                <span class="stat-number">42</span>
                <span class="stat-label">Users Online</span>
            </div>
        </div>
        
        <div class="stat-card">
            <div class="stat-icon">📊</div>
            <div class="stat-info">
                <span class="stat-number">1523</span>
                <span class="stat-label">Requests Today</span>
            </div>
        </div>
        
        <div class="stat-card">
            <div class="stat-icon">⚡</div>
            <div class="stat-info">
                <span class="stat-number">99.9%%</span>
                <span class="stat-label">Uptime</span>
            </div>
        </div>
        
        <div class="stat-card">
            <div class="stat-icon">⏱️</div>
            <div class="stat-info">
                <span class="stat-number">12ms</span>
                <span class="stat-label">Response Time</span>
            </div>
        </div>
    </div>
    
    <div class="stats-footer">
        <p><small>Last updated: %s</small></p>
    </div>
</div>

<style>
.stats-live {
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    overflow: hidden;
}

.stats-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1px;
    background: #dee2e6;
}

.stat-card {
    background: #fff;
    padding: 1.5rem;
    display: flex;
    align-items: center;
    gap: 1rem;
}

.stat-icon {
    font-size: 2rem;
}

.stat-number {
    display: block;
    font-size: 1.5rem;
    font-weight: bold;
    color: #007bff;
}

.stat-label {
    display: block;
    font-size: 0.875rem;
    color: #6c757d;
}

.stats-footer {
    background: #f8f9fa;
    padding: 0.75rem 1.5rem;
    text-align: center;
    border-top: 1px solid #dee2e6;
}
</style>
`, time.Now().Format("15:04:05"))
		return ctx.HTML(http.StatusOK, htmlContent)
	})

	// Product details
	api.GET("/products/:id", func(ctx *lokstra.Context) error {
		productID := ctx.GetPathParam("id")

		// Simulate product lookup
		product := map[string]any{
			"id":          productID,
			"name":        "Product " + productID,
			"description": "Detailed information about product " + productID,
			"price":       "$99.99",
			"in_stock":    true,
			"rating":      4.5,
			"reviews": []map[string]any{
				{"user": "John D.", "rating": 5, "comment": "Excellent product!"},
				{"user": "Jane S.", "rating": 4, "comment": "Very good, would recommend."},
			},
		}

		return ctx.Ok(product)
	})

	// ===== Form Handling =====

	type contactFormRequest struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}

	// Contact form submission
	api.POST("/contact", func(ctx *lokstra.Context, req *contactFormRequest) error {
		// Validate required fields
		if req.Name == "duplicate" || req.Email == "" || req.Message == "" {
			return ctx.HTML(http.StatusOK, `
<div class="contact-response">
    <div class="alert alert-danger">
        <h4>❌ Error</h4>
        <p>Duplicate Name detected</p>
        <p>Please check your input and try again.</p>
    </div>
</div>

<style>
.contact-response {
    margin-top: 1rem;
}

.alert {
    padding: 1rem;
    border-radius: 8px;
    border: 1px solid;
}

.alert h4 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
}

.alert p {
    margin: 0.5rem 0 0 0;
}

.alert-danger {
    background: #f8d7da;
    border-color: #f5c6cb;
    color: #721c24;
}
</style>
<script>
setTimeout(function() {
  var el = document.querySelector('.contact-response');
  if (el) el.style.display = 'none';
}, 3000);
</script>
			`)
		}

		// Simulate processing
		time.Sleep(500 * time.Millisecond)

		// Return success message
		return ctx.HTML(http.StatusOK, fmt.Sprintf(`
<div class="contact-response">
    <div class="alert alert-success">
        <h4>✅ Message Sent Successfully!</h4>
        <p>Thank you for your message! We'll get back to you soon.</p>
        <p><strong>Thank you, %s!</strong> We'll get back to you soon.</p>
    </div>
</div>

<style>
.contact-response {
    margin-top: 1rem;
}

.alert {
    padding: 1rem;
    border-radius: 8px;
    border: 1px solid;
}

.alert h4 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
}

.alert p {
    margin: 0.5rem 0 0 0;
}

.alert-success {
    background: #d4edda;
    border-color: #c3e6cb;
    color: #155724;
}
</style>
<script>
setTimeout(function() {
  var el = document.querySelector('.contact-response');
  if (el) el.style.display = 'none';
}, 5000);
</script>
		`, req.Name))
	})

	// ===== Search Functionality =====

	api.GET("/search", func(ctx *lokstra.Context) error {
		query := ctx.GetQueryParam("q")

		if query == "" {
			return ctx.HTML(http.StatusOK, `
<div class="search-empty">
    <p>Enter a search term to get started</p>
</div>

<style>
.search-empty {
    text-align: center;
    padding: 2rem;
    color: #6c757d;
    background: #f8f9fa;
    border-radius: 8px;
}
</style>
			`)
		}

		// Simulate search results
		results := []map[string]any{
			{"title": "HTMX Basics", "description": "Learn the fundamentals of HTMX", "url": "/docs/htmx-basics"},
			{"title": "Lokstra Framework", "description": "Getting started with Lokstra", "url": "/docs/lokstra"},
			{"title": "Dynamic Content", "description": "Building dynamic web pages", "url": "/docs/dynamic"},
		}

		// Filter results based on query
		filteredResults := []map[string]any{}
		for _, result := range results {
			title := result["title"].(string)
			description := result["description"].(string)
			if containsIgnoreCase(title, query) || containsIgnoreCase(description, query) {
				filteredResults = append(filteredResults, result)
			}
		}

		if len(filteredResults) == 0 {
			return ctx.HTML(http.StatusOK, fmt.Sprintf(`
<div class="search-empty">
    <p>No results found for "%s"</p>
    <p><small>Try different keywords or check your spelling</small></p>
</div>

<style>
.search-empty {
    text-align: center;
    padding: 2rem;
    color: #6c757d;
    background: #f8f9fa;
    border-radius: 8px;
}
</style>
			`, query))
		}

		var htmlBuilder strings.Builder
		htmlBuilder.WriteString(fmt.Sprintf(`
<div class="search-results-list">
    <div class="search-header">
        <h4>Search Results for "%s"</h4>
        <span class="result-count">%d results found</span>
    </div>
`, query, len(filteredResults)))

		for _, result := range filteredResults {
			htmlBuilder.WriteString(fmt.Sprintf(`
    <div class="search-result-item">
        <h5><a href="%s">%s</a></h5>
        <p>%s</p>
        <small class="result-url">%s</small>
    </div>
`, result["url"].(string), result["title"].(string), result["description"].(string), result["url"].(string)))
		}

		htmlBuilder.WriteString(`
</div>

<style>
.search-results-list {
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    overflow: hidden;
}

.search-header {
    background: #f8f9fa;
    padding: 1rem;
    border-bottom: 1px solid #dee2e6;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.search-header h4 {
    margin: 0;
    font-size: 1rem;
}

.result-count {
    font-size: 0.875rem;
    color: #6c757d;
}

.search-result-item {
    padding: 1rem;
    border-bottom: 1px solid #dee2e6;
}

.search-result-item:last-child {
    border-bottom: none;
}

.search-result-item h5 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
}

.search-result-item a {
    color: #007bff;
    text-decoration: none;
}

.search-result-item a:hover {
    text-decoration: underline;
}

.search-result-item p {
    margin: 0 0 0.5rem 0;
    color: #6c757d;
}

.result-url {
    color: #28a745;
    font-size: 0.75rem;
}
</style>
`)

		return ctx.HTML(http.StatusOK, htmlBuilder.String())
	})

	// ===== Activity Feed =====

	// Live activity feed
	api.GET("/activity/live", func(ctx *lokstra.Context) error {
		// Simulate live activity data
		activities := []map[string]any{
			{
				"id":        1,
				"type":      "user_login",
				"user":      "john_doe",
				"message":   "John Doe logged in",
				"timestamp": time.Now().Add(-2 * time.Minute).Format("15:04:05"),
				"icon":      "🔑",
			},
			{
				"id":        2,
				"type":      "content_view",
				"user":      "jane_smith",
				"message":   "Jane Smith viewed 'HTMX Tutorial'",
				"timestamp": time.Now().Add(-5 * time.Minute).Format("15:04:05"),
				"icon":      "👁️",
			},
			{
				"id":        3,
				"type":      "content_like",
				"user":      "bob_wilson",
				"message":   "Bob Wilson liked 'Go Best Practices'",
				"timestamp": time.Now().Add(-8 * time.Minute).Format("15:04:05"),
				"icon":      "❤️",
			},
			{
				"id":        4,
				"type":      "user_register",
				"user":      "alice_brown",
				"message":   "Alice Brown registered",
				"timestamp": time.Now().Add(-12 * time.Minute).Format("15:04:05"),
				"icon":      "🆕",
			},
			{
				"id":        5,
				"type":      "content_comment",
				"user":      "charlie_green",
				"message":   "Charlie Green commented on 'Microservices Guide'",
				"timestamp": time.Now().Add(-15 * time.Minute).Format("15:04:05"),
				"icon":      "💬",
			},
		}

		var htmlBuilder strings.Builder
		htmlBuilder.WriteString(`
<div class="activity-list">
`)

		for _, activity := range activities {
			htmlBuilder.WriteString(fmt.Sprintf(`
    <div class="activity-item">
        <div class="activity-icon">%s</div>
        <div class="activity-content">
            <div class="activity-message">%s</div>
            <div class="activity-timestamp">%s</div>
        </div>
    </div>
`, activity["icon"].(string), activity["message"].(string), activity["timestamp"].(string)))
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
</div>
<div class="activity-summary">
    <p><small>%d activities • Last updated: %s</small></p>
</div>

<style>
.activity-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.activity-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
}

.activity-icon {
    font-size: 1.5rem;
    flex-shrink: 0;
}

.activity-content {
    flex: 1;
}

.activity-message {
    font-weight: 500;
    margin-bottom: 0.5rem;
}

.activity-timestamp {
    font-size: 0.875rem;
    color: #6c757d;
}

.activity-summary {
    margin-top: 1rem;
    text-align: center;
    padding-top: 1rem;
    border-top: 1px solid #dee2e6;
}
</style>
`, len(activities), time.Now().Format("15:04:05")))

		return ctx.HTML(http.StatusOK, htmlBuilder.String())
	})

	// Recent activity (non-live)
	api.GET("/activity/recent", func(ctx *lokstra.Context) error {
		// Simulate recent activity data
		activities := []map[string]any{
			{
				"id":        10,
				"type":      "content_publish",
				"user":      "admin",
				"message":   "New article 'Advanced HTMX Patterns' published",
				"timestamp": time.Now().Add(-1 * time.Hour).Format("15:04:05"),
				"icon":      "📝",
			},
			{
				"id":        11,
				"type":      "user_achievement",
				"user":      "developer123",
				"message":   "Developer123 earned 'HTMX Expert' badge",
				"timestamp": time.Now().Add(-2 * time.Hour).Format("15:04:05"),
				"icon":      "🏆",
			},
			{
				"id":        12,
				"type":      "system_update",
				"user":      "system",
				"message":   "System maintenance completed successfully",
				"timestamp": time.Now().Add(-3 * time.Hour).Format("15:04:05"),
				"icon":      "⚙️",
			},
		}

		var htmlBuilder strings.Builder
		htmlBuilder.WriteString(`
<div class="activity-list">
`)

		for _, activity := range activities {
			htmlBuilder.WriteString(fmt.Sprintf(`
    <div class="activity-item">
        <div class="activity-icon">%s</div>
        <div class="activity-content">
            <div class="activity-message">%s</div>
            <div class="activity-timestamp">%s</div>
        </div>
    </div>
`, activity["icon"].(string), activity["message"].(string), activity["timestamp"].(string)))
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
</div>
<div class="activity-summary">
    <p><small>%d activities • Last updated: %s</small></p>
</div>

<style>
.activity-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.activity-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
}

.activity-icon {
    font-size: 1.5rem;
    flex-shrink: 0;
}

.activity-content {
    flex: 1;
}

.activity-message {
    font-weight: 500;
    margin-bottom: 0.5rem;
}

.activity-timestamp {
    font-size: 0.875rem;
    color: #6c757d;
}

.activity-summary {
    margin-top: 1rem;
    text-align: center;
    padding-top: 1rem;
    border-top: 1px solid #dee2e6;
}
</style>
`, len(activities), time.Now().Format("15:04:05")))

		return ctx.HTML(http.StatusOK, htmlBuilder.String())
	})

	// ===== Additional HTMX Endpoints =====

	// Featured content categories
	api.GET("/content/articles", func(ctx *lokstra.Context) error {
		return ctx.HTML(http.StatusOK, `
<div class="content-items">
    <div class="content-item">
        <h4>Understanding HTMX Basics</h4>
        <p>Learn the fundamentals of HTMX for dynamic web development.</p>
        <span class="content-meta">Published 2 days ago • 5 min read</span>
    </div>
    <div class="content-item">
        <h4>Go Web Development with Lokstra</h4>
        <p>Building modern web applications with the Lokstra framework.</p>
        <span class="content-meta">Published 1 week ago • 8 min read</span>
    </div>
    <div class="content-item">
        <h4>Advanced HTMX Patterns</h4>
        <p>Explore advanced techniques for building complex interactive UIs.</p>
        <span class="content-meta">Published 3 days ago • 12 min read</span>
    </div>
</div>

<style>
.content-items {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.content-item {
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 1.5rem;
}

.content-item h4 {
    margin: 0 0 0.5rem 0;
    color: #007bff;
}

.content-item p {
    margin: 0 0 1rem 0;
    color: #6c757d;
}

.content-meta {
    font-size: 0.875rem;
    color: #adb5bd;
}
</style>
		`)
	})

	api.GET("/content/videos", func(ctx *lokstra.Context) error {
		return ctx.HTML(http.StatusOK, `
<div class="content-items">
    <div class="content-item">
        <h4>📹 HTMX Tutorial Series</h4>
        <p>Complete video course covering HTMX from basics to advanced topics.</p>
        <span class="content-meta">Duration: 45 min • 1.2k views</span>
    </div>
    <div class="content-item">
        <h4>📹 Building with Lokstra</h4>
        <p>Step-by-step guide to creating a web application with Lokstra framework.</p>
        <span class="content-meta">Duration: 32 min • 890 views</span>
    </div>
</div>

<style>
.content-items {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.content-item {
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 1.5rem;
}

.content-item h4 {
    margin: 0 0 0.5rem 0;
    color: #007bff;
}

.content-item p {
    margin: 0 0 1rem 0;
    color: #6c757d;
}

.content-meta {
    font-size: 0.875rem;
    color: #adb5bd;
}
</style>
		`)
	})

	api.GET("/content/tutorials", func(ctx *lokstra.Context) error {
		return ctx.HTML(http.StatusOK, `
<div class="content-items">
    <div class="content-item">
        <h4>🛠️ Quick Start Guide</h4>
        <p>Get up and running with HTMX and Lokstra in 10 minutes.</p>
        <span class="content-meta">Beginner • 10 steps</span>
    </div>
    <div class="content-item">
        <h4>🛠️ Form Handling Tutorial</h4>
        <p>Learn to handle forms with HTMX for better user experience.</p>
        <span class="content-meta">Intermediate • 15 steps</span>
    </div>
    <div class="content-item">
        <h4>🛠️ Real-time Updates</h4>
        <p>Implement live updates and real-time features with HTMX.</p>
        <span class="content-meta">Advanced • 20 steps</span>
    </div>
</div>

<style>
.content-items {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.content-item {
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 1.5rem;
}

.content-item h4 {
    margin: 0 0 0.5rem 0;
    color: #007bff;
}

.content-item p {
    margin: 0 0 1rem 0;
    color: #6c757d;
}

.content-meta {
    font-size: 0.875rem;
    color: #adb5bd;
}
</style>
		`)
	})

	api.GET("/content/news", func(ctx *lokstra.Context) error {
		return ctx.HTML(http.StatusOK, `
<div class="content-items">
    <div class="content-item">
        <h4>📰 Lokstra v2.0 Released</h4>
        <p>Major update brings enhanced HTMX integration and performance improvements.</p>
        <span class="content-meta">2 hours ago • Framework News</span>
    </div>
    <div class="content-item">
        <h4>📰 HTMX 2.0 Beta Available</h4>
        <p>New features and improvements in the latest HTMX beta release.</p>
        <span class="content-meta">1 day ago • Technology News</span>
    </div>
</div>

<style>
.content-items {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.content-item {
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 1.5rem;
}

.content-item h4 {
    margin: 0 0 0.5rem 0;
    color: #007bff;
}

.content-item p {
    margin: 0 0 1rem 0;
    color: #6c757d;
}

.content-meta {
    font-size: 0.875rem;
    color: #adb5bd;
}
</style>
		`)
	})

	// Recommendations
	api.GET("/recommendations", func(ctx *lokstra.Context) error {
		return ctx.HTML(http.StatusOK, `
<div class="recommendations">
    <h4>Recommended for You</h4>
    <div class="recommendation-list">
        <div class="recommendation-item">
            <span class="rec-badge">🎯 For You</span>
            <h5>Advanced HTMX Patterns</h5>
            <p>Based on your interest in web development</p>
        </div>
        <div class="recommendation-item">
            <span class="rec-badge">📈 Trending</span>
            <h5>Go Performance Optimization</h5>
            <p>Popular among developers like you</p>
        </div>
        <div class="recommendation-item">
            <span class="rec-badge">⭐ Highly Rated</span>
            <h5>Modern Web Architecture</h5>
            <p>Matches your skill level</p>
        </div>
    </div>
</div>

<style>
.recommendations {
    background: #fff;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 1.5rem;
}

.recommendations h4 {
    margin: 0 0 1rem 0;
    color: #007bff;
}

.recommendation-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.recommendation-item {
    background: #f8f9fa;
    border: 1px solid #dee2e6;
    border-radius: 6px;
    padding: 1rem;
    position: relative;
}

.rec-badge {
    background: #007bff;
    color: white;
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    position: absolute;
    top: -8px;
    right: 8px;
}

.recommendation-item h5 {
    margin: 0 0 0.5rem 0;
}

.recommendation-item p {
    margin: 0;
    color: #6c757d;
    font-size: 0.875rem;
}
</style>
		`)
	})
}
