# Lokstra Framework – Milestone & Roadmap

Lokstra is a backend framework for Go that is flexible, scalable, and structured. Below are the milestones and development roadmap for Lokstra, organized into phases.

---

## 🧱 Milestone v0.1 – Foundation Stability (Current Focus)

**Goal:** The framework is usable for real services, with a stable and maintainable structure.

### Features:

* Modular Server, App, and RouterEngine
* RequestContext + binding + response helpers
* Middleware: global, group, and handler levels
* Service registry + basic lifecycle hooks (`OnStart`, `OnStop`)
* Service config via YAML + ENV override
* Routing engine based on httpRouter
* Examples in `/cmd/examples/*`
* LoggerService, DBPoolService, RedisWorkerService *(required)*
* SimpleWorkerService *(required)*
* Multi-port App (multiple apps in a single binary)
* Graceful shutdown per service

---

## 🚀 Milestone v0.2 – Production Readiness

**Goal:** The framework is ready for small to medium scale production environments.

### Features:

* HealthCheckService (readiness & liveness)
* MetricsService (Prometheus-ready)
* JWTAuthService + AuthMiddleware
* RecoveryMiddleware, CORS, RequestLogger
* Standardized validation and error responses
* Sample validation + response struct
* Multi-tenant support (`TenantManagementService`)
* Unit test coverage minimum 80%
* Dockerfile + sample Docker Compose

---

## ⚙️ Milestone v0.3 – Developer Experience & Documentation

**Goal:** New developers can easily learn and be productive with Lokstra.

### Features:

* Complete documentation (README, architecture, tutorials)
* CLI helper (`lokstra init`, `lokstra new service`)
* Plugin/snippet support for VSCode *(optional)*
* Form response builder (for frontend integration)
* Logging with field-based API

---

## 🌐 Milestone v0.4 – Integration Friendly & Custom Service

**Goal:** Lokstra is ready to be integrated with real-world systems and external services.

### Features:

* WebSocketService (pub/sub, async command)
* EmailService (SMTP sender)
* ExternalServiceManager (3rd party API support)
* Job Worker Service (SimpleWorker)
* CLI job runner (`lokstra run jobname`)
* Extended hooks: `OnConfigReload`, `OnTenantChanged`

---

## 🔁 Milestone v1.0 – Public Stable Release

**Goal:** The framework is stable and ready for public use.

### Criteria:

* Project structure is stable (no breaking changes after this point)
* Full documentation (README, CONTRIBUTING, CODE\_OF\_CONDUCT)
* Real-world examples (todo-api, user-auth-api, multitenant-api)
* Test coverage >90%
* Apache 2.0 license
* Go module release `v1.0.0`
* RedisWorkerService and SimpleWorkerService implemented and documented

---

## 🧠 Future Roadmap (Post v1.0)

* Hot-reload Router & Service (lokstra-runtime)
* Swagger/OpenAPI integration
* OpenTelemetry & distributed tracing
* Modular RBAC & permission mapping
* Refine.dev frontend template
* Premium plugin system (Lokstra Premium)
* Multi-language response codes + i18n middleware
* Internal RPC system (`lokstra-call`)
* Plugin support: async queue, retry, afterResponse hooks

---

For contributions or roadmap discussions, please open an issue or submit a pull request on GitHub. Thank you for supporting Lokstra!
