# Dokumentasi AI Agent untuk Lokstra Framework

## 📋 Ringkasan

Dokumentasi ini dibuat untuk membantu AI Agent (GitHub Copilot, Claude, ChatGPT, dll) memahami dan menggunakan Lokstra Framework dengan efektif saat membantu programmer.

## 📁 File-File yang Dibuat

### 1. **AI Agent Guide** (Komprehensif)
**Lokasi:** `docs/AI-AGENT-GUIDE.md`

Dokumentasi lengkap dan terstruktur yang mencakup:
- Framework overview (Router vs Framework mode)
- Core concepts (Router, App, Service, Repository, Domain)
- Quick start patterns (dengan dan tanpa DI)
- Router patterns (29+ handler signatures)
- Service patterns (Factory, DI, Lazy loading)
- Configuration YAML (schema lengkap)
- Annotation system (@RouterService, @Route, @Inject)
- Middleware usage (built-in dan custom)
- Dependency injection (LazyLoad, Cached)
- Project structure templates (3 levels)
- Common patterns & idioms
- Troubleshooting guide
- Quick reference table

**Target:** AI Agent yang ingin pemahaman mendalam tentang Lokstra

---

### 2. **Quick Reference** (Cheatsheet)
**Lokasi:** `docs/QUICK-REFERENCE.md`

Referensi cepat berisi:
- Installation commands
- Common imports
- Router basics (dengan contoh)
- Handler signatures (7 variasi umum)
- Service patterns (factory, registration, lazy loading)
- Domain models (dengan validation tags)
- Configuration YAML (minimal dan complete)
- Annotations (dengan contoh lengkap)
- Middleware (6 built-in + custom)
- Response helpers
- Validation tags reference
- Environment variables
- Project structures
- CLI commands
- Common issues & fixes
- Template comparison table

**Target:** Lookup cepat untuk pattern-pattern umum

---

### 3. **GitHub Copilot Instructions**
**Lokasi:** `.github/copilot-instructions.md`

File khusus untuk GitHub Copilot yang berisi:
- Framework overview ringkas
- Link ke dokumentasi lengkap
- Common patterns (5 pattern utama)
- Project structure (3 template)
- Common imports
- Middleware order best practice
- Response helpers
- CLI commands
- Template recommendations
- Best practices (5 poin kunci)
- When suggesting code (3 guidelines)

**Target:** GitHub Copilot saat bekerja di repository Lokstra

---

### 4. **Copilot Context File**
**Lokasi:** `.copilot`

File konteks minimal untuk AI assistants:
- About framework (ringkas)
- Essential reading (3 dokumen)
- Key concepts (6 poin)
- Quick examples (2 mode)
- When helping users (5 guidelines)
- Resources (links)

**Target:** Memberikan konteks cepat ke AI assistant

---

## 🎯 Cara Penggunaan

### Untuk AI Agent (Copilot, Claude, ChatGPT):

1. **Baca AI Agent Guide** (`docs/AI-AGENT-GUIDE.md`) untuk pemahaman komprehensif
2. **Gunakan Quick Reference** (`docs/QUICK-REFERENCE.md`) untuk lookup cepat
3. **Ikuti Best Practices** yang tercantum di semua dokumentasi

### Untuk Programmer:

1. **Tambahkan di README project:**
   ```markdown
   ## For AI Assistants
   
   This project uses Lokstra Framework. AI assistants should read:
   - [AI Agent Guide](https://primadi.github.io/lokstra/AI-AGENT-GUIDE)
   - [Quick Reference](https://primadi.github.io/lokstra/QUICK-REFERENCE)
   ```

2. **Gunakan dengan GitHub Copilot:**
   - Copilot akan otomatis membaca `.github/copilot-instructions.md`
   - Tidak perlu konfigurasi tambahan

3. **Gunakan dengan AI lain:**
   - Share link ke AI Agent Guide
   - Atau copy-paste section yang relevan

---

## 📚 Konten Utama yang Dicakup

### 1. **Router Patterns**
- 29+ handler signatures
- Group routing
- Middleware application
- Path parameters
- Query parameters
- Request body binding
- Validation

### 2. **Service Patterns**
- Factory pattern
- Service registration
- Lazy loading
- Cached dependencies
- Remote services (microservices)
- Type-safe DI

### 3. **Configuration**
- YAML schema
- Service definitions
- Middleware definitions
- Router definitions
- Deployment configurations
- Multi-environment setup

### 4. **Annotations**
- @RouterService
- @Route (HTTP methods + paths)
- @Inject
- Code generation (`lokstra autogen`)

### 5. **Middleware**
- Built-in middleware (6 jenis)
- Custom middleware
- Middleware order
- YAML configuration

### 6. **Project Structures**
- Simple router (learning)
- Medium system (DDD)
- Enterprise modular (bounded contexts)
- Enterprise annotations

### 7. **Best Practices**
- Error handling
- Validation
- Domain-driven design
- Type-safe DI
- Deployment patterns

---

## 🔗 Integrasi dengan Dokumentasi Utama

Dokumentasi AI Agent terintegrasi dengan:

1. **README.md:**
   - Link ke AI Agent Guide
   - Link ke Quick Reference

2. **docs/index.md:**
   - Card untuk AI Agent Guide
   - Card untuk Quick Reference
   - Link di Community & Support

3. **GitHub Pages:**
   - https://primadi.github.io/lokstra/AI-AGENT-GUIDE
   - https://primadi.github.io/lokstra/QUICK-REFERENCE

---

## 🎨 Fitur Khusus

### 1. Contoh Code Lengkap
Semua contoh code mencakup:
- ✅ Import statements
- ✅ Error handling
- ✅ Validation tags
- ✅ Config YAML (jika diperlukan)
- ✅ Complete working code

### 2. Multiple Learning Paths
- **Path 1:** Router mode (simple, no DI)
- **Path 2:** Framework mode (DI + config)
- **Path 3:** Enterprise (annotations + modular)

### 3. Template Recommendations
AI Agent diarahkan untuk menyarankan template berdasarkan:
- Project scale (small, medium, large)
- Team size
- Complexity level
- Architecture pattern

### 4. Troubleshooting Guide
Common issues dengan solusi:
- Service not found
- Import cycle
- Handler not recognized
- Validation not working
- Annotation code not generated
- Middleware not applied

---

## 🚀 Next Steps untuk AI Agent

Saat membantu programmer dengan Lokstra:

1. **Assess Project Scale:**
   - Small/learning → Router mode
   - Medium/production → Framework mode + DDD
   - Enterprise → Modular + annotations

2. **Suggest Appropriate Template:**
   - `01_router/01_router_only` - Learning
   - `02_app_framework/01_medium_system` - Production
   - `02_app_framework/03_enterprise_router_service` - Enterprise

3. **Provide Complete Solution:**
   - Full imports
   - Error handling
   - Validation tags
   - Config YAML (if framework mode)
   - Project structure

4. **Follow Best Practices:**
   - Domain → Repository → Service
   - Type-safe DI
   - Lazy loading
   - Proper middleware order

---

## 📊 Statistik Dokumentasi

- **AI Agent Guide:** ~1,200 baris (komprehensif)
- **Quick Reference:** ~600 baris (ringkas)
- **Copilot Instructions:** ~200 baris (fokus)
- **Copilot Context:** ~80 baris (minimal)
- **Total Examples:** 50+ code snippets
- **Coverage:** Router, Service, Config, Annotations, Middleware, DI

---

## ✅ Checklist Kualitas

- [x] Contoh code lengkap dan runnable
- [x] Error handling di semua contoh
- [x] Validation tags included
- [x] Config YAML examples
- [x] Import statements
- [x] Best practices documented
- [x] Troubleshooting guide
- [x] Quick reference table
- [x] Multiple learning paths
- [x] Template recommendations
- [x] Integration with main docs
- [x] Links to full documentation

---

## 🌐 Resources

- **AI Agent Guide:** https://primadi.github.io/lokstra/AI-AGENT-GUIDE
- **Quick Reference:** https://primadi.github.io/lokstra/QUICK-REFERENCE
- **Full Documentation:** https://primadi.github.io/lokstra/
- **GitHub:** https://github.com/primadi/lokstra
- **Templates:** https://github.com/primadi/lokstra/tree/dev2/project_templates

---

**Status:** ✅ Complete  
**Created:** November 12, 2025  
**Author:** AI Assistant  
**Purpose:** Enable AI agents to effectively help programmers using Lokstra Framework
