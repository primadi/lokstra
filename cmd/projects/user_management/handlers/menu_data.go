package handlers

// MenuItem represents a single menu item in the sidebar
type MenuItem struct {
	ID         string     `json:"id"` // id unik untuk menu
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	Icon       string     `json:"icon"`
	IconRule   bool       `json:"icon_rule"`
	CSSClass   string     `json:"css_class"`
	IsDropdown bool       `json:"is_dropdown"`
	IsOpen     bool       `json:"is_open"`
	SubItems   []MenuItem `json:"sub_items,omitempty"`
}

// SidebarData contains all data needed for sidebar rendering
type SidebarData struct {
	MenuItems []MenuItem `json:"menu_items"`
}

// getMenuData returns the menu configuration based on current page
func getMenuData(currentPage string) SidebarData {
	// Base CSS classes
	defaultClass := "flex items-center p-3 text-gray-300 rounded-lg hover:bg-gray-700 transition-colors"
	activeClass := "flex items-center p-3 text-gray-300 rounded-lg bg-gray-700 border border-gray-600"
	activeSubClass := "bg-gray-700"

	menuItems := []MenuItem{
		// Dashboard - Heroicons: squares-2x2
		{
			ID:         "menu-dashboard",
			Title:      "Dashboard",
			URL:        "/dashboard",
			Icon:       "M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z",
			IconRule:   true,
			CSSClass:   defaultClass,
			IsDropdown: false,
		},
		// User Management (Dropdown) - Heroicons: users
		{
			ID:         "menu-user-management",
			Title:      "User Management",
			URL:        "",
			Icon:       "M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z",
			IconRule:   true,
			CSSClass:   "",
			IsDropdown: true,
			IsOpen:     false,
			SubItems: []MenuItem{
				{
					ID:       "menu-all-users",
					Title:    "All Users",
					URL:      "/users",
					Icon:     "M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z",
					IconRule: true,
					CSSClass: "",
				},
				{
					ID:       "menu-add-user",
					Title:    "Add New User",
					URL:      "/users/new",
					Icon:     "M18 7.5v3m0 0v3m0-3h3m-3 0h-3m-2.25-4.125a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zM3 19.235v-.11a6.375 6.375 0 0112.75 0v.109A12.318 12.318 0 019.374 21c-2.331 0-4.512-.645-6.374-1.766z",
					IconRule: false,
					CSSClass: "",
				},
				{
					ID:       "menu-user-stats",
					Title:    "User Statistics",
					URL:      "/api/v1/admin/users/stats",
					Icon:     "M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z",
					IconRule: false,
					CSSClass: "",
				},
			},
		},
		// Reports (Dropdown) - Heroicons: chart-bar-square
		{
			ID:         "menu-reports",
			Title:      "Reports",
			URL:        "",
			Icon:       "M7.5 14.25v2.25m3-4.5v4.5m3-6.75v6.75m3-9v9M6 20.25h12A2.25 2.25 0 0020.25 18V6A2.25 2.25 0 0018 3.75H6A2.25 2.25 0 003.75 6v12A2.25 2.25 0 006 20.25z",
			IconRule:   false,
			CSSClass:   "",
			IsDropdown: true,
			IsOpen:     false,
			SubItems: []MenuItem{
				{
					ID:       "menu-user-reports",
					Title:    "User Reports",
					URL:      "/reports/users",
					Icon:     "M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z",
					IconRule: false,
					CSSClass: "",
				},
				{
					ID:       "menu-activity-reports",
					Title:    "Activity Reports",
					URL:      "/reports/activity",
					Icon:     "M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5M9 11.25v1.5M12 9v3.75m3-6v6",
					IconRule: false,
					CSSClass: "",
				},
			},
		},
		// Settings - Heroicons: cog-6-tooth
		{
			ID:         "menu-settings",
			Title:      "Settings",
			URL:        "/settings",
			Icon:       "M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z M15 12a3 3 0 11-6 0 3 3 0 016 0z",
			IconRule:   true,
			CSSClass:   defaultClass,
			IsDropdown: false,
		},
		// Health Check - Heroicons: heart
		{
			ID:         "menu-health-check",
			Title:      "Health Check",
			URL:        "/health",
			Icon:       "M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z",
			IconRule:   false,
			CSSClass:   defaultClass,
			IsDropdown: false,
		},
	}

	// Apply active states based on currentPage
	for i := range menuItems {
		if menuItems[i].Title == "Dashboard" && currentPage == "dashboard" {
			menuItems[i].CSSClass = activeClass
		}

		if menuItems[i].Title == "User Management" {
			// Open user management dropdown if on user-related page
			if currentPage == "users" || currentPage == "user_form" {
				menuItems[i].IsOpen = true

				// Apply active state to sub-items
				for j := range menuItems[i].SubItems {
					if (menuItems[i].SubItems[j].URL == "/users" && currentPage == "users") ||
						(menuItems[i].SubItems[j].URL == "/users/new" && currentPage == "user_form") {
						menuItems[i].SubItems[j].CSSClass = activeSubClass
					}
				}
			}
		}

		if menuItems[i].Title == "Settings" && currentPage == "settings" {
			menuItems[i].CSSClass = activeClass
		}

		if menuItems[i].Title == "Health Check" && currentPage == "health" {
			menuItems[i].CSSClass = activeClass
		}

		// Handle Reports dropdown
		if menuItems[i].Title == "Reports" {
			if currentPage == "reports" || currentPage == "user_reports" || currentPage == "activity_reports" {
				menuItems[i].IsOpen = true

				// Apply active state to sub-items
				for j := range menuItems[i].SubItems {
					if (menuItems[i].SubItems[j].URL == "/reports/users" && currentPage == "user_reports") ||
						(menuItems[i].SubItems[j].URL == "/reports/activity" && currentPage == "activity_reports") {
						menuItems[i].SubItems[j].CSSClass = activeSubClass
					}
				}
			}
		}
	}

	return SidebarData{
		MenuItems: menuItems,
	}
}
