package handlers

import (
	"github.com/primadi/lokstra"
)

// Example of how to create new pages using the layout utility system

// CreateSimplePageHandler demonstrates basic usage of SmartPageHandler
func CreateSimplePageHandler() lokstra.HandlerFunc {
	config := DashboardLayout
	config.Title = "Simple Demo Page"
	config.CurrentPage = "dashboard"

	return SmartPageHandler(func(c *lokstra.Context) (string, error) {
		content := `
			<div class="bg-gray-800 rounded-lg shadow-lg border border-gray-700 p-6">
				<h2 class="text-2xl font-bold text-gray-100 mb-4">Simple Demo Page</h2>
				<p class="text-gray-300 mb-4">
					This page demonstrates how easy it is to create new pages using the layout utility system.
				</p>
				<div class="bg-blue-900 border border-blue-700 rounded p-4 mb-4">
					<h3 class="text-blue-200 font-semibold mb-2">Features:</h3>
					<ul class="text-blue-300 space-y-1">
						<li>• Automatic full page vs HTMX partial detection</li>
						<li>• Consistent styling and layout</li>
						<li>• Easy content focus - just write HTML content</li>
						<li>• Configurable title, scripts, and styles</li>
					</ul>
				</div>
				<button hx-get="/api/content/users" 
						hx-target="#main-content"
						class="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded transition-colors">
					Go to Users (HTMX)
				</button>
				<a href="/users" class="ml-3 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded transition-colors inline-block">
					Go to Users (Direct)
				</a>
			</div>
		`
		return content, nil
	}, config)
}

// CreateAdvancedPageHandler demonstrates consistent behavior with page-specific assets in content
func CreateAdvancedPageHandler() lokstra.HandlerFunc {
	config := PageConfig{
		Title:       "Advanced Demo",
		CurrentPage: "dashboard",
	}

	return SmartPageHandler(func(c *lokstra.Context) (string, error) {
		// Page-specific assets should be in content for consistency
		content := `
			<!-- Page-specific styles for this page only -->
			<style>
				.custom-card {
					background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
					border-radius: 12px;
					padding: 2rem;
					color: white;
				}
				.pulse-animation {
					animation: pulse 2s infinite;
				}
			</style>
			
			<div class="space-y-6">
				<div class="custom-card animate__animated animate__fadeInDown">
					<h2 class="text-2xl font-bold mb-4">Advanced Demo Page</h2>
					<p class="opacity-90">
						This page demonstrates consistent behavior - page-specific assets are in content,
						so they work the same for both full page load and HTMX partial load.
					</p>
				</div>
				
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div class="bg-gray-800 rounded-lg p-6 border border-gray-700">
						<h3 class="text-xl font-semibold text-gray-100 mb-3">Chart Demo</h3>
						<canvas id="demoChart" width="400" height="200"></canvas>
					</div>
					
					<div class="bg-gray-800 rounded-lg p-6 border border-gray-700">
						<h3 class="text-xl font-semibold text-gray-100 mb-3">Animation Demo</h3>
						<div class="pulse-animation bg-blue-600 rounded-full w-16 h-16 mx-auto mb-4"></div>
						<p class="text-gray-300 text-center">Custom CSS animations working!</p>
					</div>
				</div>
			</div>

			<script>
				// Chart.js demo (only runs when Chart.js is loaded)
				if (typeof Chart !== 'undefined') {
					const ctx = document.getElementById('demoChart');
					if (ctx) {
						new Chart(ctx, {
							type: 'bar',
							data: {
								labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
								datasets: [{
									label: 'Demo Data',
									data: [12, 19, 3, 5, 2, 3],
									backgroundColor: '#3B82F6'
								}]
							},
							options: {
								responsive: true,
								plugins: {
									legend: {
										labels: {
											color: '#E5E7EB'
										}
									}
								},
								scales: {
									y: {
										ticks: {
											color: '#E5E7EB'
										},
										grid: {
											color: '#374151'
										}
									},
									x: {
										ticks: {
											color: '#E5E7EB'
										},
										grid: {
											color: '#374151'
										}
									}
								}
							}
						});
					}
				}
			</script>
		`
		return content, nil
	}, config)
}

// CreateContentOnlyHandler demonstrates simple content-only handler for HTMX endpoints
func CreateContentOnlyHandler() lokstra.HandlerFunc {
	config := PageConfig{
		Title:       "Content Only",
		CurrentPage: "dashboard",
	}

	return SmartPageHandler(func(c *lokstra.Context) (string, error) {
		content := `
			<div class="bg-gray-800 rounded-lg shadow-lg border border-gray-700 p-6">
				<h2 class="text-xl font-bold text-gray-100 mb-4">Content-Only Handler</h2>
				<p class="text-gray-300 mb-4">
					This handler demonstrates consistent behavior between full page and HTMX partial loads.
				</p>
				<div class="bg-yellow-900 border border-yellow-700 rounded p-4">
					<p class="text-yellow-200">
						💡 Same content, same behavior - whether loaded via full page or HTMX.
					</p>
				</div>
			</div>
		`
		return content, nil
	}, config)
}
