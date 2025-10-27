# Lokstra Documentation Layouts

This folder contains custom Jekyll layouts for the Lokstra documentation site.

## Available Layouts

### 1. `default.html` (Landing Page)
Used for the home page and marketing content.

**Features:**
- Full-width content
- No sidebar
- Clean navigation bar
- Feature cards and hero sections

**Usage:**
```yaml
---
layout: default
title: Home
---
```

### 2. `docs.html` (Documentation Pages)
Used for all documentation, guides, and API reference pages.

**Features:**
- **Left sidebar** with navigation tree
- **Breadcrumb navigation** at the top
- **Sticky header** and sidebar
- **Mobile-responsive** with hamburger menu
- **Active page highlighting** in sidebar

**Usage:**
```yaml
---
layout: docs
title: Your Page Title
---
```

## Sidebar Navigation Structure

The sidebar in `docs.html` is organized into 3 sections:

1. **Introduction** - Overview, architecture, examples
2. **Essentials** - Getting started guides
3. **API Reference** - Technical documentation

To add new pages to the sidebar, edit `_layouts/docs.html` and add entries to the appropriate section.

## How to Add New Documentation

1. Create your markdown file in the appropriate folder
2. Add front matter at the top:
   ```yaml
   ---
   layout: docs
   title: My New Page
   ---
   ```
3. Add a link to the sidebar in `_layouts/docs.html`:
   ```html
   <li><a href="{{ site.baseurl }}/path/to/your-page/">Your Page Title</a></li>
   ```

## Customization

All styles are embedded in the layout files. To customize:

- **Colors**: Edit CSS variables in `:root` section
- **Sidebar width**: Change `--sidebar-width` variable
- **Fonts**: Edit `font-size` and `font-family` properties
- **Spacing**: Adjust padding/margin values

## Mobile Behavior

On screens smaller than 1024px:
- Sidebar slides in from the left
- Hamburger button (☰) appears in bottom-right
- Tap outside sidebar to close
- Breadcrumb navigation remains visible
