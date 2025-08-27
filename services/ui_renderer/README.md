# 🎨 Lokstra UI Renderer Service

**Memberdayakan Programmer Go untuk Membuat Aplikasi Web Lengkap dengan Mudah!**

## 🎯 Visi

Lokstra UI Renderer Service dirancang khusus untuk **programmer Go yang ingin membuat aplikasi web lengkap tanpa harus menguasai frontend development**. Dengan menggunakan konfigurasi YAML yang sederhana, Anda dapat menghasilkan UI modern yang menggunakan:

- **Preline UI** - Komponen UI modern dan responsif
- **Tailwind CSS** - Styling yang fleksibel dan powerful  
- **HTMX** - Interaksi server-side yang smooth tanpa JavaScript kompleks
- **Alpine.js** - Reaktivitas client-side yang ringan

## ✨ Fitur Utama

### 🚀 **Deklaratif & Mudah**
```yaml
# Cukup tulis YAML seperti ini:
forms:
  create_user:
    title: "Create New User"
    fields:
      - name: "email"
        type: "email" 
        label: "Email Address"
        required: true
```

### 📱 **Responsive & Modern**
- Layout sidebar, top navigation, dan minimal
- Komponen UI yang mengikuti design system terbaru
- Mobile-first responsive design
- Dark mode support (opsional)

### ⚡ **HTMX Integration**
```yaml
htmx:
  endpoint: "/api/users"
  method: "post"
  target: "#user-list"
  swap: "afterbegin"
```

### 🎪 **Alpine.js Reactivity**
```yaml
alpine:
  data: "{ showModal: false, selectedUser: null }"
  show: "showModal"
  click: "selectedUser = user; showModal = true"
```

## 🏗️ Arsitektur

```
UI Renderer Service
├── 📄 Templates (Go html/template)
│   ├── layout.html
│   ├── form.html
│   ├── table.html
│   └── components.html
├── ⚙️ Service Implementation
│   ├── module.go
│   └── service.go
└── 🔧 Service API Interface
    └── ui_renderer.go
```

## 🎮 Cara Penggunaan

### 1. **Instalasi Service**

```go
import (
    "github.com/primadi/lokstra/services/ui_renderer"
    "github.com/primadi/lokstra/serviceapi"
)

// Setup UI Renderer
uiConfig := &ui_renderer.Config{
    TemplateDir: "./templates",
    StaticDir:   "./static",
}

uiService := ui_renderer.NewService(uiConfig)
```

### 2. **Render Application Layout**

```go
appConfig := &serviceapi.AppConfig{
    Title:  "User Management System",
    Layout: "sidebar",
    Menu: serviceapi.MenuConfig{
        Items: []serviceapi.MenuItem{
            {
                Label: "Dashboard",
                URL:   "/dashboard", 
                Icon:  "fas fa-tachometer-alt",
                Active: true,
            },
            {
                Label: "Users",
                URL:   "/users",
                Icon:  "fas fa-users",
            },
        },
    },
}

html, err := uiService.RenderApp(ctx, appConfig)
```

### 3. **Create Forms dengan HTMX**

```go
formConfig := &serviceapi.FormConfig{
    Title: "Create New User",
    HTMX: serviceapi.HTMXConfig{
        URL:    "/api/users",
        Method: "post",
        Target: "#user-list",
        Swap:   "afterbegin",
    },
    Fields: []serviceapi.FieldConfig{
        {
            Name:        "email",
            Type:        "email",
            Label:       "Email Address", 
            Required:    true,
            Placeholder: "user@example.com",
        },
        {
            Name:  "tenant_id",
            Type:  "select",
            Label: "Tenant",
            Options: []serviceapi.OptionConfig{
                {Value: "1", Label: "Acme Corp"},
                {Value: "2", Label: "TechStart Inc"},
            },
        },
    },
}

formHTML, err := uiService.RenderForm(ctx, formConfig)
```

### 4. **Create Tables dengan Filtering**

```go
listConfig := &serviceapi.ListConfig{
    Title:       "Users Management",
    Description: "Manage system users and permissions",
}

tableHTML, err := uiService.RenderList(ctx, listConfig, userData)
```

### 5. **Components (Modal, Alert, etc)**

```go
// Modal
modalProps := map[string]interface{}{
    "title": "User Details",
    "show":  false,
}
modalHTML, err := uiService.RenderComponent(ctx, "modal", modalProps)

// Alert
alertProps := map[string]interface{}{
    "type":    "success",
    "title":   "Success!",
    "message": "User created successfully",
}
alertHTML, err := uiService.RenderComponent(ctx, "alert", alertProps)
```

## 📋 Komponen yang Tersedia

### 🖼️ **Layouts**
- **Sidebar Layout** - Perfect untuk admin panels
- **Top Navigation** - Great untuk landing pages  
- **Minimal Layout** - Clean untuk forms

### 📝 **Forms**
- Text, Email, Password inputs
- Select, Radio, Checkbox
- Textarea, File upload
- Validation dengan error messages
- HTMX integration untuk submission
- Alpine.js untuk interactivity

### 📊 **Tables/Lists**
- Sortable columns
- Search & filtering
- Pagination
- Row actions (Edit, Delete, View)
- Bulk actions
- Empty states
- HTMX untuk real-time updates

### 🎭 **Components**
- **Modal** - Untuk detail dan forms
- **Alert** - Success, error, warning notifications
- **Breadcrumb** - Navigation path
- **Card** - Content containers
- **Badges** - Status indicators

## 🎨 Design System

### 🎯 **Preline UI Components**
Service ini menggunakan Preline UI yang memberikan:
- 500+ komponen siap pakai
- Consistent design language
- Accessibility compliant
- Professional appearance

### 🎨 **Tailwind CSS Styling**
```html
<!-- Generated output menggunakan Tailwind classes -->
<form class="max-w-2xl mx-auto bg-white shadow-lg rounded-lg p-6">
  <input class="block w-full rounded-md border-gray-300 shadow-sm 
                focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm">
</form>
```

### ⚡ **HTMX Integration**
```html
<!-- Automatic HTMX attributes -->
<form hx-post="/api/users" hx-target="#user-list" hx-swap="afterbegin">
  <!-- Form fields -->
</form>
```

### 🎪 **Alpine.js Reactivity**
```html
<!-- Alpine.js attributes untuk interactivity -->
<div x-data="{ showModal: false }">
  <button @click="showModal = true">Open Modal</button>
  <div x-show="showModal">Modal Content</div>
</div>
```

## 🚀 Contoh Lengkap: User Management

```go
package main

import (
    "context"
    "github.com/primadi/lokstra/services/ui_renderer"
    "github.com/primadi/lokstra/serviceapi"
)

func main() {
    // Initialize service
    ui := ui_renderer.NewService(&ui_renderer.Config{
        TemplateDir: "./templates",
    })
    
    ctx := context.Background()
    
    // 1. App Layout
    app := &serviceapi.AppConfig{
        Title:  "User Management",
        Layout: "sidebar",
        Menu: serviceapi.MenuConfig{
            Items: []serviceapi.MenuItem{
                {Label: "Users", URL: "/users", Icon: "fas fa-users"},
                {Label: "Settings", URL: "/settings", Icon: "fas fa-cog"},
            },
        },
    }
    appHTML, _ := ui.RenderApp(ctx, app)
    
    // 2. User Form
    form := &serviceapi.FormConfig{
        Title: "Create User",
        HTMX: serviceapi.HTMXConfig{
            URL: "/api/users", Method: "post", Target: "#users",
        },
        Fields: []serviceapi.FieldConfig{
            {Name: "name", Type: "text", Label: "Name", Required: true},
            {Name: "email", Type: "email", Label: "Email", Required: true},
        },
    }
    formHTML, _ := ui.RenderForm(ctx, form)
    
    // 3. Users Table
    list := &serviceapi.ListConfig{
        Title: "All Users",
    }
    userData := []map[string]interface{}{
        {"name": "John Doe", "email": "john@example.com"},
        {"name": "Jane Smith", "email": "jane@example.com"},
    }
    listHTML, _ := ui.RenderList(ctx, list, userData)
    
    // Output siap untuk HTTP response!
    fmt.Println("✅ Complete UI rendered!")
}
```

## 🎯 Use Cases

### 👥 **Admin Panels**
- User management systems
- Content management
- E-commerce admin
- Analytics dashboards

### 📊 **Business Applications**
- CRM systems
- Project management tools
- Inventory management
- Financial applications

### 🏢 **Enterprise Systems**
- HR management
- Document management
- Workflow systems
- Reporting tools

## 🔧 Advanced Configuration

### 🎨 **Custom CSS Classes**
```go
formConfig := &serviceapi.FormConfig{
    CSS: &serviceapi.FormCSS{
        Container: "max-w-4xl mx-auto bg-white shadow-xl rounded-lg p-8",
        Grid:      "grid grid-cols-1 md:grid-cols-2 gap-6",
        Actions:   "flex justify-end space-x-4 pt-8 border-t",
    },
}
```

### ⚡ **HTMX Advanced Features**
```go
htmxConfig := serviceapi.HTMXConfig{
    URL:       "/api/users/search",
    Method:    "get", 
    Target:    "#results",
    Trigger:   "keyup changed delay:500ms",
    Indicator: "#loading",
    Confirm:   "Are you sure?",
}
```

### 🎪 **Alpine.js Integration**
```go
alpineConfig := serviceapi.AlpineConfig{
    Data:  "{ users: [], loading: false }",
    Init:  "fetchUsers()",
    Show:  "!loading",
    Click: "selectUser(user.id)",
}
```

## 📁 Project Structure

```
your-project/
├── cmd/
│   └── your-app/
│       └── main.go
├── templates/          # UI Renderer templates
│   ├── layout.html
│   ├── form.html
│   ├── table.html
│   └── components.html
├── static/            # CSS, JS, images
│   ├── css/
│   ├── js/
│   └── images/
└── config/
    └── ui_config.yaml # Optional YAML configs
```

## 🎁 Template Examples

### 📱 **Responsive Layout**
```html
<!DOCTYPE html>
<html class="h-full bg-gray-50">
<head>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://preline.co/assets/css/main.min.css" rel="stylesheet">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body class="h-full">
    <!-- Sidebar layout dengan responsive design -->
</body>
</html>
```

### 📝 **Smart Forms**
```html
<form hx-post="/api/users" hx-target="#user-list">
    <input type="email" 
           class="block w-full rounded-md border-gray-300 
                  focus:border-indigo-500 focus:ring-indigo-500"
           required>
    <button type="submit" 
            class="bg-indigo-600 hover:bg-indigo-700 text-white 
                   px-4 py-2 rounded-md">
        Create User
    </button>
</form>
```

## 🚀 Getting Started

1. **Install Dependencies**
```bash
go mod tidy
```

2. **Copy Templates**
```bash
cp -r services/ui_renderer/templates ./templates
```

3. **Run Example**
```bash
go run cmd/examples/ui_renderer_demo/main_simple.go
```

4. **Start Building!**
```go
// Mulai dengan app layout
app := &serviceapi.AppConfig{
    Title: "My App",
    Layout: "sidebar",
}

// Tambahkan forms
form := &serviceapi.FormConfig{
    Title: "Create Item",
    Fields: []serviceapi.FieldConfig{
        {Name: "name", Type: "text", Label: "Name"},
    },
}

// Render dan gunakan!
html, _ := ui.RenderApp(ctx, app)
```

## 🤝 Contributing

Kami sangat welcome kontribusi untuk:
- Template components baru
- Layout variations
- HTMX integrations
- Alpine.js helpers
- Documentation improvements

## 📄 License

MIT License - Gunakan dengan bebas untuk project komersial maupun personal.

---

## 💡 **Mengapa UI Renderer Service?**

### 🎯 **Untuk Go Developers**
- **No Frontend Expertise Required** - Fokus pada business logic Go Anda
- **Rapid Prototyping** - Dari konsep ke UI dalam hitungan menit
- **Maintainable** - YAML configuration yang mudah dibaca dan diubah
- **Type Safe** - Interface Go yang jelas dengan compile-time checking

### 🚀 **Modern Web Technologies**
- **Preline UI** - Professional component library
- **Tailwind CSS** - Utility-first styling yang powerful
- **HTMX** - Server-side rendering dengan SPA-like UX
- **Alpine.js** - Lightweight reactivity tanpa kompleksitas

### 🏢 **Production Ready**
- **Performance** - Server-side rendering yang cepat
- **SEO Friendly** - HTML yang proper untuk search engines  
- **Accessibility** - WCAG compliant components
- **Security** - Server-side validation dan CSRF protection

---

**🎉 Mulai membuat aplikasi web yang amazing dengan Go + Lokstra UI Renderer Service!**
