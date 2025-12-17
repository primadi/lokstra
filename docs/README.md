# Lokstra Documentation

Welcome to the Lokstra Framework documentation directory.

## 🏗 Project Overview

Lokstra is a **modern Go web framework with declarative service management** that can be used in two primary ways:

- **Router Mode** – like Gin / Echo / Chi, focused on HTTP routing & middleware.
- **Application Framework Mode** – like NestJS / Spring Boot, with dependency injection, auto-generated routers, and config-driven deployment.

High-level repository structure:

- **`cmd/lokstra`** – Lokstra CLI:  
  - Create new projects from templates (`lokstra new ...`).  
  - Run code generation for annotations / enterprise templates (`lokstra autogen ...`).  
  - Basic docs are in `cmd/lokstra/README.md` (for full parameter list, see the source code).

- **`common`** – Shared utilities used across projects:
  - **`api_client`** – helpers for calling remote services over HTTP.
  - **`cast`** – type conversion helpers (e.g. string → int).
  - **`customtype`** – custom types like `date`, `datetime`, `decimal`.
  - **`dbpool_manager`** – helpers for database connection pooling.
  - **`json`** – JSON encoding/decoding helpers.
  - **`logger`** – logging utilities based on `slog` with readable handlers.
  - **`response_writer`** – helpers for buffering and manipulating responses.
  - **`syncmap`** – concurrency-safe maps with a nicer API.
  - **`utils`** – generic helpers (string, slice, map, hash password, etc.).
  - **`validator`** – validation integration using struct tags.

- **`core`** – Lokstra framework core:
  - Routing (`core/router`), request/response handling, and route description.
  - Deployment & configuration loading (`core/deploy`).
  - Service & server abstractions (`core/service`, `core/server`, `core/app`).
  - Annotation and code generation support (`core/annotation`).

- **`docs`** – This documentation folder, published to GitHub Pages (`https://primadi.github.io/lokstra/`) using Jekyll.

- **`lokstra_handler`** – Handler helpers:
  - `mount_reverse_proxy`, `mount_static`, `mount_spa` to easily mount reverse proxy, static files, and SPA.

- **`lokstra_init`** – Initialization helpers for Lokstra:
  - Application bootstrap, config loading, service registration, etc.

- **`lokstra_registry`** – Lokstra Dependency Injection / Service Registry:
  - Register service types, factories, and service configurations.

- **`middleware`** – Built-in middleware:
  - Recovery, request logger, slow request logger, gzip compression, CORS, and more.

- **`project_templates`** – Running project examples and templates for the CLI:
  - **`01_router`** – Lokstra as a **Router**:
    - `01_router_only` – pure router, good for learning routing & CRUD.
    - `02_single_app` – single app server with graceful shutdown.
    - `03_multi_app` – multiple apps (e.g. admin + public API).
  - **`02_app_framework`** – Lokstra as an **Application Framework**:
    - `01_enterprise_router_service` – enterprise modular app with annotations, generated routers, and multiple topologies.
    - `02_sync_config` – configuration sync / infrastructure service example (db pools, migrations, sync map).
    - `03_tenant_management` – tenant management service using a Postgres-backed store.

- **`serviceapi`** – Interfaces for built-in services (dbpool, redis, email, metrics, etc.) that you can implement or mock.

- **`services`** – Implementations of built-in services:
  - Database pool, Redis, SMTP email sender, Prometheus metrics, sync config, and more.

To see end-to-end application examples:

- **Lokstra as Router**: see `project_templates/01_router/*`.
- **Lokstra as Application Framework**: see `project_templates/02_app_framework/*`.

## 📁 Documentation Structure

### Main Documentation (GitHub Pages)

- **[index.md](./index.md)** - Homepage and overview
- **[00-introduction/](./00-introduction/)** - Getting started, examples, architecture
- **[01-router-guide/](./01-router-guide/)** - Router mode (like Echo, Gin, Chi)
- **[02-framework-guide/](./02-framework-guide/)** - Framework mode (like NestJS, Spring Boot)
- **[03-api-reference/](./03-api-reference/)** - Complete API reference

### AI Assistant Documentation

- **[AI-AGENT-GUIDE.md](./AI-AGENT-GUIDE.md)** - Comprehensive guide for AI agents (Copilot, Claude, ChatGPT)
- **[QUICK-REFERENCE.md](./QUICK-REFERENCE.md)** - Fast lookup cheatsheet
- **[AI-DOCUMENTATION-SUMMARY.md](./AI-DOCUMENTATION-SUMMARY.md)** - Overview of AI documentation

### Other Documentation

- **[ROADMAP.md](./ROADMAP.md)** - Future plans and features

## 🤖 For AI Assistants

If you're an AI assistant helping a programmer with Lokstra Framework:

1. **Start here:** [AI-AGENT-GUIDE.md](./AI-AGENT-GUIDE.md)
2. **Quick lookup:** [QUICK-REFERENCE.md](./QUICK-REFERENCE.md)
3. **Full docs:** https://primadi.github.io/lokstra/

## 🌐 Online Documentation

**Live Site:** https://primadi.github.io/lokstra/

The documentation is built using Jekyll and hosted on GitHub Pages.

## 📝 Contributing to Documentation

### Local Development

1. Install Jekyll:
   ```bash
   gem install bundler jekyll
   ```

2. Serve locally:
   ```bash
   cd docs
   jekyll serve
   ```

3. Open browser:
   ```
   http://localhost:4000/lokstra/
   ```

### File Organization

- **Pages:** Markdown files (`.md`)
- **Layouts:** `./_layouts/` directory
- **Assets:** `./assets/` directory (images, CSS, etc.)
- **Config:** `./_config.yml`

### Adding New Pages

1. Create `.md` file with front matter:
   ```yaml
   ---
   layout: default
   title: Your Page Title
   description: Page description
   ---
   ```

2. Add content in Markdown

3. Update navigation if needed

### Adding Code Examples

Use fenced code blocks with language:

````markdown
```go
package main

func main() {
    // Your code here
}
```
````

### Schema Files

- **[schema/](./schema/)** - JSON schema for config.yaml validation

## 🔗 Related Files

### Root Directory

- **[../.github/copilot-instructions.md](../.github/copilot-instructions.md)** - GitHub Copilot specific instructions
- **[../.copilot](../.copilot)** - AI assistant context file
- **[../README.md](../README.md)** - Main repository README

### Project Templates

- **[../project_templates/](../project_templates/)** - Starter templates for new projects

## 📊 Documentation Statistics

- **Main Pages:** 50+ pages
- **Code Examples:** 100+ snippets
- **AI Documentation:** 1,800+ lines
- **Templates:** 6 project templates
- **Languages:** English

## 🛠️ Tools & Technologies

- **Static Site Generator:** Jekyll
- **Hosting:** GitHub Pages
- **Domain:** primadi.github.io/lokstra
- **Theme:** Custom (based on GitHub's minimal theme)
- **Markdown Processor:** Kramdown
- **Syntax Highlighting:** Rouge

## 📖 Documentation Sections

### 1. Introduction (00-introduction/)
- Quick start guide
- Examples (Hello World, JSON API, CRUD, Multi-deployment)
- Architecture overview
- Code vs Config comparison

### 2. Router Guide (01-router-guide/)
- Basic routing
- Handler signatures
- Middleware
- Groups and versioning
- Request/Response handling

### 3. Framework Guide (02-framework-guide/)
- Service management
- Dependency injection
- Configuration (YAML)
- Deployment patterns
- Comparisons (vs NestJS, vs Spring Boot)

### 4. API Reference (03-api-reference/)
- Registry API
- Router registration
- Service registration
- Configuration schema
- Deployment patterns

### 5. AI Assistant Guide
- Comprehensive AI guide
- Quick reference cheatsheet
- Best practices
- Troubleshooting

## 🎯 Target Audiences

1. **Go Developers** - Learning Lokstra
2. **Enterprise Teams** - Building scalable apps
3. **AI Assistants** - Helping programmers
4. **Contributors** - Improving Lokstra

## 🚀 Quick Links

- **Live Docs:** https://primadi.github.io/lokstra/
- **GitHub:** https://github.com/primadi/lokstra
- **Issues:** https://github.com/primadi/lokstra/issues
- **Discussions:** https://github.com/primadi/lokstra/discussions

## 📝 License

Documentation is licensed under Apache 2.0 License.

---

**Last Updated:** November 12, 2025
