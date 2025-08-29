package web_render

// SidebarComponentData is a reusable struct for sidebar rendering
// Can be used by any project for sidebar menu
// Example usage: pass SidebarComponentData to template

type SidebarComponentData struct {
	MenuItems []SidebarMenuItem
}

type SidebarMenuItem struct {
	ID         string
	Title      string
	URL        string
	Icon       string
	IconRule   bool
	CSSClass   string
	IsDropdown bool
	IsOpen     bool
	SubItems   []SidebarMenuItem
}

// Helper to create sidebar data from custom menu
func NewSidebarComponent(menuItems []SidebarMenuItem) SidebarComponentData {
	return SidebarComponentData{MenuItems: menuItems}
}
