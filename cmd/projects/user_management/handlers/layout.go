package handlers

import (
	"fmt"

	"github.com/primadi/lokstra"
)

// PageContent holds all page data for consistent behavior
type PageContent struct {
	HTML        string            // Main content HTML
	Title       string            // Page title (for browser tab and meta)
	CurrentPage string            // Current page identifier (for sidebar active state)
	Scripts     []string          // Page-specific scripts
	Styles      []string          // Page-specific styles
	CustomCSS   string            // Page-specific custom CSS
	MetaTags    map[string]string // Page-specific meta tags
	SidebarData interface{}       // Custom sidebar data if needed
}

// PageContentFunc is a function that returns complete page content
type PageContentFunc func(*lokstra.Context) (*PageContent, error)

// RenderFullPage renders a complete HTML page with layout
func RenderFullPage(pageContent *PageContent) string {
	sidebarHTML := getSidebarHTML(pageContent.CurrentPage)

	// Global meta tags + page-specific meta tags
	metaTags := `<meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">`

	// Add page-specific meta tags
	for name, content := range pageContent.MetaTags {
		metaTags += fmt.Sprintf(`
    <meta name="%s" content="%s">`, name, content)
	}

	// Global scripts + page-specific scripts
	scripts := `<script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            darkMode: 'class',
        }
    </script>
    <script src="https://unpkg.com/htmx.org@1.9.0"></script>
    <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
    <script>
        // Simple debouncing to prevent rapid navigation
        let navigationInProgress = false;
        let navigationDebounceTimeout = null;
        
        document.addEventListener('DOMContentLoaded', function() {
            // Configure HTMX
            htmx.config.timeout = 10000;
            htmx.config.historyCacheSize = 3;
            htmx.config.defaultSwapStyle = 'innerHTML';
            htmx.config.globalViewTransitions = true;
            
            // Set global defaults for all HTMX requests
            document.body.addEventListener('htmx:configRequest', function(evt) {
                if (!evt.detail.headers['hx-indicator']) {
                    evt.detail.indicator = '#loading-indicator';
                }
                
                if (evt.detail.verb === 'get' && 
                    evt.detail.target && 
                    evt.detail.target.id === 'main-content' &&
                    !evt.detail.headers['hx-push-url']) {
                    evt.detail.headers['hx-push-url'] = evt.detail.path;
                }
            });
            
            // Navigation protection
            htmx.on('htmx:beforeRequest', function(evt) {
                if (navigationInProgress) {
                    console.log('Navigation blocked - request in progress');
                    evt.preventDefault();
                    return false;
                }
                
                if (navigationDebounceTimeout) {
                    clearTimeout(navigationDebounceTimeout);
                }
                
                navigationInProgress = true;
                console.log('Navigation started:', evt.detail.pathInfo.requestPath);
            });
            
            htmx.on('htmx:afterRequest', function(evt) {
                navigationDebounceTimeout = setTimeout(() => {
                    navigationInProgress = false;
                    console.log('Navigation completed');
                    updateSidebarActiveMenu();
                }, 200);
            });
            
            htmx.on('htmx:responseError', function(evt) {
                navigationInProgress = false;
                console.warn('Navigation error:', evt.detail);
            });
            
            htmx.on('htmx:timeout', function(evt) {
                navigationInProgress = false;
                console.warn('Navigation timeout');
            });
            
            // Initial sidebar update
            updateSidebarActiveMenu();
            
            // Listen for browser history changes
            window.addEventListener('popstate', function(evt) {
                console.log('Popstate event - updating sidebar menu');
                updateSidebarActiveMenu();
            });
            
            // Auto-setup navigation elements
            setupNavigationElements();
        });
        
        function setupNavigationElements() {
            document.querySelectorAll('.nav-page[hx-get]').forEach(function(element) {
                const href = element.getAttribute('hx-get');
                
                if (!element.hasAttribute('hx-target')) {
                    element.setAttribute('hx-target', '#main-content');
                }
                
                if (!element.hasAttribute('hx-swap')) {
                    element.setAttribute('hx-swap', 'innerHTML');
                }
                
                if (!element.hasAttribute('hx-indicator')) {
                    element.setAttribute('hx-indicator', '#loading-indicator');
                }
                
                if (!element.hasAttribute('hx-push-url')) {
                    element.setAttribute('hx-push-url', href);
                }
                
                console.log('Auto-setup navigation for:', href);
            });
        }
        
        function updateSidebarActiveMenu() {
            const currentPath = window.location.pathname;
            console.log('Updating sidebar for path:', currentPath);
            
            document.querySelectorAll('.sidebar-nav-link').forEach(link => {
                link.classList.remove('bg-gray-700', 'border', 'border-gray-600');
                link.classList.add('text-gray-300');
                link.classList.remove('text-white');
            });
            
            let foundMatch = false;
            document.querySelectorAll('.sidebar-nav-link').forEach(link => {
                const href = link.getAttribute('href');
                if (href) {
                    if (currentPath === href) {
                        link.classList.add('bg-gray-700', 'border', 'border-gray-600', 'text-white');
                        link.classList.remove('text-gray-300');
                        foundMatch = true;
                        
                        const dropdown = link.closest('li[x-data]');
                        if (dropdown && dropdown._x_dataStack) {
                            dropdown._x_dataStack[0].menuOpen = true;
                        }
                    }
                    else if (!foundMatch && href !== '/' && href !== '/users' && currentPath.startsWith(href)) {
                        link.classList.add('bg-gray-700', 'border', 'border-gray-600', 'text-white');
                        link.classList.remove('text-gray-300');
                        
                        const dropdown = link.closest('li[x-data]');
                        if (dropdown && dropdown._x_dataStack) {
                            dropdown._x_dataStack[0].menuOpen = true;
                        }
                    }
                }
            });
        }
    </script>`

	// Add page-specific scripts
	for _, script := range pageContent.Scripts {
		scripts += fmt.Sprintf(`
    <script src="%s"></script>`, script)
	}

	// Global styles + page-specific styles
	styles := `<style>
        /* HTMX Loading Indicator */
        .htmx-indicator {
            display: none;
        }
        .htmx-request .htmx-indicator {
            display: flex !important;
        }
        
        /* Global loading state for HTMX requests */
        .htmx-request #loading-indicator {
            display: flex !important;
        }
        
        /* Custom navigation classes for automatic HTMX behavior */
        .nav-page {
            /* Automatically adds navigation behavior */
        }
        
        /* Smooth transitions for content updates */
        #main-content {
            transition: opacity 0.2s ease-in-out;
        }
        .htmx-request #main-content {
            opacity: 0.7;
        }
        
        /* Navigation active state animation */
        .nav-item {
            transition: all 0.2s ease;
        }
        .nav-item:hover {
            transform: translateX(4px);
        }
    </style>`

	// Add page-specific styles
	for _, style := range pageContent.Styles {
		styles += fmt.Sprintf(`
    <link rel="stylesheet" href="%s">`, style)
	}

	// Add page-specific custom CSS
	if pageContent.CustomCSS != "" {
		styles += fmt.Sprintf(`
    <style>
        %s
    </style>`, pageContent.CustomCSS)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" class="dark">
<head>
    %s
    <title>%s - User Management</title>
    %s
    %s
</head>
<body class="bg-gray-900 text-gray-100" x-data="{ sidebarOpen: false }">
    <div class="min-h-screen flex">
        %s
        
        <!-- Mobile sidebar backdrop -->
        <div x-show="sidebarOpen" 
             @click="sidebarOpen = false"
             x-transition:enter="transition-opacity ease-linear duration-300"
             x-transition:enter-start="opacity-0"
             x-transition:enter-end="opacity-100"
             x-transition:leave="transition-opacity ease-linear duration-300"
             x-transition:leave-start="opacity-100"
             x-transition:leave-end="opacity-0"
             class="fixed inset-0 z-40 bg-gray-600 bg-opacity-75 lg:hidden">
        </div>
        
        <!-- Main Content -->
        <div class="flex-1 lg:ml-0">
            <header class="bg-gray-800 shadow-lg border-b border-gray-700">
                <div class="px-6 py-4 flex items-center justify-between">
                    <div class="flex items-center">
                        <button @click="sidebarOpen = true" class="lg:hidden text-gray-400 hover:text-white mr-4">
                            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
                            </svg>
                        </button>
                        <h1 class="text-2xl font-bold text-white">User Management</h1>
                    </div>
                    <div class="flex items-center space-x-4">
                        <span class="text-gray-300">%s</span>
                        <div class="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center">
                            <span class="text-white text-sm font-medium">A</span>
                        </div>
                    </div>
                </div>
            </header>
            
            <!-- Loading Indicator - Outside main content to prevent HTMX replacement -->
            <div id="loading-indicator" class="htmx-indicator fixed top-4 right-4 z-50">
                <div class="bg-blue-600 text-white px-4 py-2 rounded-lg shadow-lg flex items-center space-x-2">
                    <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span>Loading...</span>
                </div>
            </div>
            
            <main class="p-6" id="main-content">
                %s
            </main>
        </div>
    </div>
</body>
</html>`, metaTags, pageContent.Title, scripts, styles, sidebarHTML, pageContent.Title, pageContent.HTML)
}

// RenderPartialContent renders just the content for HTMX requests
// WITH page-specific assets for consistent behavior
func RenderPartialContent(pageContent *PageContent) string {
	content := pageContent.HTML

	// Add page-specific scripts to content for HTMX consistency
	if len(pageContent.Scripts) > 0 {
		for _, script := range pageContent.Scripts {
			content = fmt.Sprintf(`<script src="%s"></script>
%s`, script, content)
		}
	}

	// Add page-specific styles to content for HTMX consistency
	if len(pageContent.Styles) > 0 || pageContent.CustomCSS != "" {
		stylesBlock := ""
		for _, style := range pageContent.Styles {
			stylesBlock += fmt.Sprintf(`<link rel="stylesheet" href="%s">
`, style)
		}
		if pageContent.CustomCSS != "" {
			stylesBlock += fmt.Sprintf(`<style>
%s
</style>
`, pageContent.CustomCSS)
		}
		content = stylesBlock + content
	}

	return content
}

// UnifiedPageHandler creates a handler with truly consistent behavior
func UnifiedPageHandler(contentFunc PageContentFunc) lokstra.HandlerFunc {
	return func(c *lokstra.Context) error {
		// Get complete page content
		pageContent, err := contentFunc(c)
		if err != nil {
			return err
		}

		// Check if this is an HTMX request for partial content
		isHTMXRequest := c.GetHeader("HX-Request") == "true"

		if isHTMXRequest {
			// Return content WITH page-specific assets for consistency
			html := RenderPartialContent(pageContent)
			return c.HTML(200, html)
		}

		// Return full page for direct access
		fullPageHTML := RenderFullPage(pageContent)
		return c.HTML(200, fullPageHTML)
	}
}

// Legacy SmartPageHandler for backward compatibility
// Converts old contentFunc to new PageContent structure
func SmartPageHandler(contentFunc func(*lokstra.Context) (string, error), config PageConfig) lokstra.HandlerFunc {
	return UnifiedPageHandler(func(c *lokstra.Context) (*PageContent, error) {
		content, err := contentFunc(c)
		if err != nil {
			return nil, err
		}

		return &PageContent{
			HTML:        content,
			Title:       config.Title,
			CurrentPage: config.CurrentPage,
			Scripts:     []string{},
			Styles:      []string{},
			CustomCSS:   "",
			MetaTags:    map[string]string{},
			SidebarData: nil,
		}, nil
	})
}

// PageConfig for backward compatibility
type PageConfig struct {
	Title       string
	CurrentPage string
}

var (
	// Legacy configs for backward compatibility
	DashboardLayout = PageConfig{
		Title:       "Dashboard",
		CurrentPage: "dashboard",
	}

	UsersLayout = PageConfig{
		Title:       "Users",
		CurrentPage: "users",
	}

	UserFormLayout = PageConfig{
		Title:       "User Form",
		CurrentPage: "user_form",
	}

	RolesLayout = PageConfig{
		Title:       "Roles",
		CurrentPage: "roles",
	}

	SettingsLayout = PageConfig{
		Title:       "Settings",
		CurrentPage: "settings",
	}
)
