package serviceapi

import (
	"context"
	"html/template"
)

// UIRenderer interface for rendering HTML forms, lists, and layouts
type UIRenderer interface {
	// App Container Methods
	RenderApp(ctx context.Context, appConfig *AppConfig) (template.HTML, error)
	RenderMenu(ctx context.Context, menuConfig *MenuConfig) (template.HTML, error)
	RenderBreadcrumb(ctx context.Context, breadcrumb *BreadcrumbConfig) (template.HTML, error)

	// Form Rendering Methods
	RenderForm(ctx context.Context, formConfig *FormConfig) (template.HTML, error)
	RenderField(ctx context.Context, field *FieldConfig) (template.HTML, error)

	// List/Table Rendering Methods
	RenderList(ctx context.Context, listConfig *ListConfig, data interface{}) (template.HTML, error)
	RenderTable(ctx context.Context, tableConfig *TableConfig, data interface{}) (template.HTML, error)
	RenderPagination(ctx context.Context, paginationConfig *PaginationConfig) (template.HTML, error)

	// Component Methods
	RenderComponent(ctx context.Context, componentName string, props map[string]interface{}) (template.HTML, error)
	RenderModal(ctx context.Context, modalConfig *ModalConfig) (template.HTML, error)
	RenderCard(ctx context.Context, cardConfig *CardConfig) (template.HTML, error)

	// Template Methods
	ParseTemplate(templatePath string) error
	RenderTemplate(ctx context.Context, templateName string, data interface{}) (template.HTML, error)
}

// AppConfig defines the overall application layout
type AppConfig struct {
	Title       string            `yaml:"title" json:"title"`
	Description string            `yaml:"description" json:"description"`
	Theme       string            `yaml:"theme" json:"theme"`   // light, dark, auto
	Layout      string            `yaml:"layout" json:"layout"` // sidebar, topnav, minimal
	Logo        *LogoConfig       `yaml:"logo" json:"logo"`
	Menu        *MenuConfig       `yaml:"menu" json:"menu"`
	Sidebar     *SidebarConfig    `yaml:"sidebar" json:"sidebar"`
	Header      *HeaderConfig     `yaml:"header" json:"header"`
	Footer      *FooterConfig     `yaml:"footer" json:"footer"`
	Meta        map[string]string `yaml:"meta" json:"meta"`
	Scripts     []string          `yaml:"scripts" json:"scripts"`
	Styles      []string          `yaml:"styles" json:"styles"`
}

type LogoConfig struct {
	Src    string `yaml:"src" json:"src"`
	Alt    string `yaml:"alt" json:"alt"`
	Width  string `yaml:"width" json:"width"`
	Height string `yaml:"height" json:"height"`
	Link   string `yaml:"link" json:"link"`
}

// MenuConfig defines navigation menu structure
type MenuConfig struct {
	Items []MenuItem `yaml:"items" json:"items"`
	Style string     `yaml:"style" json:"style"` // horizontal, vertical, dropdown
}

type MenuItem struct {
	ID       string       `yaml:"id" json:"id"`
	Label    string       `yaml:"label" json:"label"`
	Icon     string       `yaml:"icon" json:"icon"`
	URL      string       `yaml:"url" json:"url"`
	Active   bool         `yaml:"active" json:"active"`
	Children []MenuItem   `yaml:"children" json:"children"`
	Badge    *BadgeConfig `yaml:"badge" json:"badge"`
	Divider  bool         `yaml:"divider" json:"divider"`
}

type BadgeConfig struct {
	Text  string `yaml:"text" json:"text"`
	Color string `yaml:"color" json:"color"` // primary, success, warning, danger
}

// BreadcrumbConfig defines breadcrumb navigation
type BreadcrumbConfig struct {
	Items []BreadcrumbItem `yaml:"items" json:"items"`
}

type BreadcrumbItem struct {
	Label string `yaml:"label" json:"label"`
	URL   string `yaml:"url" json:"url"`
	Icon  string `yaml:"icon" json:"icon"`
}

// SidebarConfig defines sidebar layout
type SidebarConfig struct {
	Width       string `yaml:"width" json:"width"`
	Collapsible bool   `yaml:"collapsible" json:"collapsible"`
	DefaultOpen bool   `yaml:"default_open" json:"default_open"`
}

// HeaderConfig defines header layout
type HeaderConfig struct {
	Height string `yaml:"height" json:"height"`
	Sticky bool   `yaml:"sticky" json:"sticky"`
}

// FooterConfig defines footer layout
type FooterConfig struct {
	Text    string            `yaml:"text" json:"text"`
	Links   []FooterLink      `yaml:"links" json:"links"`
	Columns []FooterColumn    `yaml:"columns" json:"columns"`
	Meta    map[string]string `yaml:"meta" json:"meta"`
}

type FooterLink struct {
	Label string `yaml:"label" json:"label"`
	URL   string `yaml:"url" json:"url"`
}

type FooterColumn struct {
	Title string       `yaml:"title" json:"title"`
	Links []FooterLink `yaml:"links" json:"links"`
}

// FormConfig defines form structure and validation
type FormConfig struct {
	ID          string            `yaml:"id" json:"id"`
	Title       string            `yaml:"title" json:"title"`
	Description string            `yaml:"description" json:"description"`
	Action      string            `yaml:"action" json:"action"`
	Method      string            `yaml:"method" json:"method"` // GET, POST, PUT, DELETE
	Fields      []FieldConfig     `yaml:"fields" json:"fields"`
	Buttons     []ButtonConfig    `yaml:"buttons" json:"buttons"`
	Layout      string            `yaml:"layout" json:"layout"` // vertical, horizontal, grid
	Columns     int               `yaml:"columns" json:"columns"`
	Validation  *ValidationConfig `yaml:"validation" json:"validation"`
	HTMX        *HTMXConfig       `yaml:"htmx" json:"htmx"`
}

// FieldConfig defines individual form field
type FieldConfig struct {
	ID           string                 `yaml:"id" json:"id"`
	Name         string                 `yaml:"name" json:"name"`
	Label        string                 `yaml:"label" json:"label"`
	Type         string                 `yaml:"type" json:"type"` // text, email, password, select, textarea, checkbox, radio, file, date, etc.
	Placeholder  string                 `yaml:"placeholder" json:"placeholder"`
	Value        interface{}            `yaml:"value" json:"value"`
	DefaultValue interface{}            `yaml:"default_value" json:"default_value"`
	Required     bool                   `yaml:"required" json:"required"`
	Disabled     bool                   `yaml:"disabled" json:"disabled"`
	ReadOnly     bool                   `yaml:"readonly" json:"readonly"`
	Hidden       bool                   `yaml:"hidden" json:"hidden"`
	Options      []OptionConfig         `yaml:"options" json:"options"` // for select, radio, checkbox
	Validation   *FieldValidationConfig `yaml:"validation" json:"validation"`
	Help         string                 `yaml:"help" json:"help"`
	Icon         string                 `yaml:"icon" json:"icon"`
	Width        string                 `yaml:"width" json:"width"` // full, half, third, quarter
	Classes      []string               `yaml:"classes" json:"classes"`
	Attributes   map[string]string      `yaml:"attributes" json:"attributes"`
	Alpine       *AlpineConfig          `yaml:"alpine" json:"alpine"`
}

type OptionConfig struct {
	Label    string      `yaml:"label" json:"label"`
	Value    interface{} `yaml:"value" json:"value"`
	Selected bool        `yaml:"selected" json:"selected"`
	Disabled bool        `yaml:"disabled" json:"disabled"`
	Icon     string      `yaml:"icon" json:"icon"`
}

type ButtonConfig struct {
	ID         string            `yaml:"id" json:"id"`
	Label      string            `yaml:"label" json:"label"`
	Type       string            `yaml:"type" json:"type"`   // submit, button, reset
	Style      string            `yaml:"style" json:"style"` // primary, secondary, success, danger, warning
	Size       string            `yaml:"size" json:"size"`   // sm, md, lg
	Icon       string            `yaml:"icon" json:"icon"`
	Disabled   bool              `yaml:"disabled" json:"disabled"`
	Loading    bool              `yaml:"loading" json:"loading"`
	Classes    []string          `yaml:"classes" json:"classes"`
	Attributes map[string]string `yaml:"attributes" json:"attributes"`
	HTMX       *HTMXConfig       `yaml:"htmx" json:"htmx"`
}

type ValidationConfig struct {
	Rules    []ValidationRule  `yaml:"rules" json:"rules"`
	Messages map[string]string `yaml:"messages" json:"messages"`
}

type FieldValidationConfig struct {
	Required  bool             `yaml:"required" json:"required"`
	MinLength int              `yaml:"min_length" json:"min_length"`
	MaxLength int              `yaml:"max_length" json:"max_length"`
	Pattern   string           `yaml:"pattern" json:"pattern"`
	Min       float64          `yaml:"min" json:"min"`
	Max       float64          `yaml:"max" json:"max"`
	Email     bool             `yaml:"email" json:"email"`
	URL       bool             `yaml:"url" json:"url"`
	Custom    []ValidationRule `yaml:"custom" json:"custom"`
}

type ValidationRule struct {
	Rule    string      `yaml:"rule" json:"rule"`
	Message string      `yaml:"message" json:"message"`
	Value   interface{} `yaml:"value" json:"value"`
}

// ListConfig defines list/table structure
type ListConfig struct {
	ID          string            `yaml:"id" json:"id"`
	Title       string            `yaml:"title" json:"title"`
	Description string            `yaml:"description" json:"description"`
	Columns     []ColumnConfig    `yaml:"columns" json:"columns"`
	Actions     []ActionConfig    `yaml:"actions" json:"actions"`
	BulkActions []ActionConfig    `yaml:"bulk_actions" json:"bulk_actions"`
	Search      *SearchConfig     `yaml:"search" json:"search"`
	Filter      *FilterConfig     `yaml:"filter" json:"filter"`
	Sort        *SortConfig       `yaml:"sort" json:"sort"`
	Pagination  *PaginationConfig `yaml:"pagination" json:"pagination"`
	Layout      string            `yaml:"layout" json:"layout"` // table, grid, list
	Responsive  bool              `yaml:"responsive" json:"responsive"`
	Striped     bool              `yaml:"striped" json:"striped"`
	Bordered    bool              `yaml:"bordered" json:"bordered"`
	Hover       bool              `yaml:"hover" json:"hover"`
	HTMX        *HTMXConfig       `yaml:"htmx" json:"htmx"`
}

type TableConfig = ListConfig // Alias for backward compatibility

type ColumnConfig struct {
	ID         string       `yaml:"id" json:"id"`
	Label      string       `yaml:"label" json:"label"`
	Field      string       `yaml:"field" json:"field"`
	Type       string       `yaml:"type" json:"type"` // text, number, date, boolean, image, link, badge, actions
	Width      string       `yaml:"width" json:"width"`
	Sortable   bool         `yaml:"sortable" json:"sortable"`
	Searchable bool         `yaml:"searchable" json:"searchable"`
	Hidden     bool         `yaml:"hidden" json:"hidden"`
	Format     string       `yaml:"format" json:"format"` // date format, number format, etc.
	Link       *LinkConfig  `yaml:"link" json:"link"`
	Badge      *BadgeConfig `yaml:"badge" json:"badge"`
	Classes    []string     `yaml:"classes" json:"classes"`
	Template   string       `yaml:"template" json:"template"` // custom template
}

type ActionConfig struct {
	ID         string            `yaml:"id" json:"id"`
	Label      string            `yaml:"label" json:"label"`
	Icon       string            `yaml:"icon" json:"icon"`
	URL        string            `yaml:"url" json:"url"`
	Method     string            `yaml:"method" json:"method"`
	Style      string            `yaml:"style" json:"style"`
	Size       string            `yaml:"size" json:"size"`
	Confirm    string            `yaml:"confirm" json:"confirm"`     // confirmation message
	Condition  string            `yaml:"condition" json:"condition"` // show condition
	Classes    []string          `yaml:"classes" json:"classes"`
	Attributes map[string]string `yaml:"attributes" json:"attributes"`
	HTMX       *HTMXConfig       `yaml:"htmx" json:"htmx"`
}

type SearchConfig struct {
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	Placeholder string   `yaml:"placeholder" json:"placeholder"`
	Fields      []string `yaml:"fields" json:"fields"` // fields to search in
	Live        bool     `yaml:"live" json:"live"`     // live search with HTMX
}

type FilterConfig struct {
	Enabled bool          `yaml:"enabled" json:"enabled"`
	Fields  []FilterField `yaml:"fields" json:"fields"`
}

type FilterField struct {
	ID      string         `yaml:"id" json:"id"`
	Label   string         `yaml:"label" json:"label"`
	Field   string         `yaml:"field" json:"field"`
	Type    string         `yaml:"type" json:"type"` // select, date_range, text, number
	Options []OptionConfig `yaml:"options" json:"options"`
}

type SortConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Default string `yaml:"default" json:"default"` // default sort field
	Order   string `yaml:"order" json:"order"`     // asc, desc
}

type PaginationConfig struct {
	Enabled   bool  `yaml:"enabled" json:"enabled"`
	PageSize  int   `yaml:"page_size" json:"page_size"`
	ShowTotal bool  `yaml:"show_total" json:"show_total"`
	ShowSize  bool  `yaml:"show_size" json:"show_size"`
	Sizes     []int `yaml:"sizes" json:"sizes"` // available page sizes
}

type LinkConfig struct {
	URL    string `yaml:"url" json:"url"`
	Target string `yaml:"target" json:"target"` // _blank, _self
	Title  string `yaml:"title" json:"title"`
}

// Modal and Card Configs
type ModalConfig struct {
	ID       string         `yaml:"id" json:"id"`
	Title    string         `yaml:"title" json:"title"`
	Size     string         `yaml:"size" json:"size"` // sm, md, lg, xl, full
	Closable bool           `yaml:"closable" json:"closable"`
	Content  interface{}    `yaml:"content" json:"content"`
	Buttons  []ButtonConfig `yaml:"buttons" json:"buttons"`
	HTMX     *HTMXConfig    `yaml:"htmx" json:"htmx"`
}

type CardConfig struct {
	ID          string         `yaml:"id" json:"id"`
	Title       string         `yaml:"title" json:"title"`
	Description string         `yaml:"description" json:"description"`
	Image       string         `yaml:"image" json:"image"`
	Content     interface{}    `yaml:"content" json:"content"`
	Actions     []ActionConfig `yaml:"actions" json:"actions"`
	Classes     []string       `yaml:"classes" json:"classes"`
}

// HTMX and Alpine.js Configs
type HTMXConfig struct {
	Get     string            `yaml:"get" json:"get"`
	Post    string            `yaml:"post" json:"post"`
	Put     string            `yaml:"put" json:"put"`
	Delete  string            `yaml:"delete" json:"delete"`
	Target  string            `yaml:"target" json:"target"`
	Swap    string            `yaml:"swap" json:"swap"`       // innerHTML, outerHTML, afterbegin, beforeend, etc.
	Trigger string            `yaml:"trigger" json:"trigger"` // click, change, keyup, etc.
	Headers map[string]string `yaml:"headers" json:"headers"`
	Params  map[string]string `yaml:"params" json:"params"`
}

type AlpineConfig struct {
	Data       map[string]interface{} `yaml:"data" json:"data"`
	Show       string                 `yaml:"show" json:"show"`
	Hide       string                 `yaml:"hide" json:"hide"`
	If         string                 `yaml:"if" json:"if"`
	For        string                 `yaml:"for" json:"for"`
	Click      string                 `yaml:"click" json:"click"`
	Change     string                 `yaml:"change" json:"change"`
	Model      string                 `yaml:"model" json:"model"`
	Text       string                 `yaml:"text" json:"text"`
	HTML       string                 `yaml:"html" json:"html"`
	Class      map[string]string      `yaml:"class" json:"class"`
	Style      map[string]string      `yaml:"style" json:"style"`
	Attributes map[string]string      `yaml:"attributes" json:"attributes"`
}
