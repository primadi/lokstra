---
layout: docs
title: Key Features
description: High-level feature overview of Lokstra.
---

## Dual Mode: Router & Framework

- **Router Mode (Track 1)** – use Lokstra like Gin/Echo/Chi.  
- **Framework Mode (Track 2)** – use Lokstra like NestJS/Spring Boot.

You can mix both in the same codebase and migrate gradually.

## Router Highlights

- Grouped routes and clean middleware.  
- 20+ handler signatures (return values, errors, context parameters).  
- Strong parameter binding & validation via struct tags.  
- Built‑in middleware: recovery, logging, slow‑request logger, gzip, cors, etc.

## Framework Highlights

- **lokstra_init** bootstrap helpers (`BootstrapAndRun`, options for config, db, migrations).  
- **YAML‑driven** config: services, db pools, deployments, servers.  
- **Annotations**:
  - `@RouterService` – define a service + router.  
  - `@Route` – map methods to HTTP endpoints.  
  - `@Inject` – inject dependencies (by name, or via `@store.*` helpers).
- **Lazy DI** – services are created on first use, type‑safe via generics.

## Deployment & Topology

- Single binary can run as **monolith** or **multiple microservices** based on YAML.  
- Configurable servers and published services (`deployments` / `servers` keys).  
- Built‑in patterns for remote services and HTTP proxies.

## Templates & Tooling

- `cmd/lokstra` CLI:
  - `lokstra new` – create projects from templates.  
  - `lokstra autogen` – generate code from annotations.
- Templates in `project_templates` show:
  - Router‑only apps.  
  - Enterprise modular apps.  
  - Tenant management, sync‑config, and more.

Use this page as a map; each feature is demonstrated in at least one template.


