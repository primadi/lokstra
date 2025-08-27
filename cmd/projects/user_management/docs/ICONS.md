# 🎨 Icon Library Documentation

## 📋 Current Icon Library: **Heroicons**

Project ini menggunakan **Heroicons** sebagai icon library utama. Heroicons adalah icon library resmi dari Tailwind CSS team yang sangat cocok dengan design system kami.

### 🌟 Mengapa Heroicons?

- ✅ **Perfect match** dengan Tailwind CSS
- ✅ **Konsisten design** - semua icon memiliki style yang seragam
- ✅ **Modern & Clean** - design minimalist yang professional
- ✅ **SVG-based** - scalable dan crisp di semua ukuran
- ✅ **Free & Open Source** - MIT License
- ✅ **Regular updates** - actively maintained

### 🔗 Resources

- **Website**: https://heroicons.com/
- **GitHub**: https://github.com/tailwindlabs/heroicons
- **Figma**: https://www.figma.com/community/file/1143911270904501801

## 📦 Current Icons Used

### Main Navigation Icons

| Menu Item | Icon Name | Heroicons Path |
|-----------|-----------|----------------|
| **Dashboard** | `squares-2x2` | Layout grid icon for overview pages |
| **User Management** | `users` | Multi-user icon for user-related features |
| **Reports** | `chart-bar-square` | Chart in square for analytics/reports |
| **Settings** | `cog-6-tooth` | Classic gear icon for configuration |
| **Health Check** | `heart` | Heart icon for system health |

### Sub-menu Icons

| Sub Item | Icon Name | Description |
|----------|-----------|-------------|
| **All Users** | `user-group` | Group of users icon |
| **Add New User** | `user-plus` | User with plus sign |
| **User Statistics** | `chart-bar` | Vertical bar chart |
| **User Reports** | `document` | Document/file icon |
| **Activity Reports** | `presentation-chart-line` | Line chart presentation |

## 🛠️ Implementation Guide

### Finding New Icons

1. Visit https://heroicons.com/
2. Search for the icon you need
3. Choose between **Outline** (24x24) or **Solid** variants
4. Copy the SVG `<path>` content

### Adding New Icons to Menu

```go
// In menu_data.go
{
    Title:    "New Menu Item",
    URL:      "/new-path",
    Icon:     "M... (paste SVG path here)",
    IconRule: false, // true for fill-rule="evenodd"
    CSSClass: "",
},
```

### Icon Sizing & Styling

Icons menggunakan Tailwind classes:
- **Main menu**: `w-5 h-5` (20px)
- **Sub menu**: `w-4 h-4` (16px)
- **Color**: `fill="currentColor"` (mengikuti text color)

## 🎯 Alternative Icon Libraries

Jika perlu icon yang tidak tersedia di Heroicons:

### 1. **Lucide Icons** (Recommended Alternative)
- **Website**: https://lucide.dev/
- **Pros**: 1000+ icons, very comprehensive
- **Style**: Similar to Heroicons, stroke-based

### 2. **Tabler Icons** (For Complex UI)
- **Website**: https://tabler-icons.io/
- **Pros**: 4000+ icons, very detailed
- **Style**: Modern, pixel-perfect

### 3. **Phosphor Icons** (Unique Style)
- **Website**: https://phosphoricons.com/
- **Pros**: Flexible weight system (Thin, Light, Regular, Bold, Fill)
- **Style**: Unique, modern geometric

### 4. **Feather Icons** (Minimalist)
- **Website**: https://feathericons.com/
- **Pros**: Super minimalist, very lightweight
- **Style**: Simple stroke-based

## 📝 Icon Usage Guidelines

### ✅ Do's
- Use consistent icon style throughout the app
- Keep icon size consistent per context (main menu vs sub-menu)
- Use semantic icons (heart for health, users for user management)
- Test icons in both light and dark themes

### ❌ Don'ts
- Don't mix different icon libraries in the same interface
- Don't use icons that are too complex for small sizes
- Don't use decorative icons for functional elements
- Don't change icon style mid-project

## 🔄 Icon Update Process

1. **Identify need** for new/different icon
2. **Search Heroicons** first - maintain consistency
3. **If not available**, consider alternatives from approved libraries
4. **Update menu_data.go** with new SVG path
5. **Test in both themes** (light/dark)
6. **Update this documentation**

---

*Last Updated: August 26, 2025*
*Icon Library: Heroicons v2.x*
