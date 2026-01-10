# Lokstra App Framework Templates

**Production-Ready Templates for Enterprise & Infrastructure Scenarios**

This folder contains **application framework templates** that demonstrate how to build
enterprise-style applications and infrastructure services using Lokstra,
including **DDD modules**, **annotation-based routers**, and **config-driven deployment**.

---

## 📂 Available Templates

### 1. [`01_enterprise_router_service`](./01_enterprise_router_service/)

**Enterprise modular application with annotation-based routers**

This template shows how to build a **large, modular application** with:

- ✅ **DDD modules** (`modules/user`, `modules/order`, `modules/shared`)
- ✅ `@EndpointService` / `@Route` annotations and **generated routers**
- ✅ Per-environment deployments (`development` vs `microservice`) via `config/deployment.yaml`
- ✅ Custom middleware: `request-logger`, `simple-auth`, `mw-test`
- ✅ New bootstrap flow using `lokstra_init.BootstrapAndRun`

**Use when**:

- You want to see the **full enterprise pattern** for Track 2 (Application Framework).  
- You need an example of **bounded contexts + annotations + auto routers**.  
- You want to understand **how to structure modules** and **configure multiple topologies**.

For details, see [`01_enterprise_router_service/README.md`](./01_enterprise_router_service/README.md)
and [`README_FLOWS.md`](./01_enterprise_router_service/README_FLOWS.md).

---

### 2. [`02_sync_config`](./02_sync_config/)

**Configuration synchronization & migrations example**

This template demonstrates how to build a **background / infrastructure service**
that uses Lokstra features such as:

- ✅ `lokstra_init.BootstrapAndRun` with:
  - `WithAnnotations`
  - `WithYAMLConfigPath`
  - `WithPgSyncMap`
  - `WithDbPoolManager`
  - `WithDbMigrations`
- ✅ Database migrations stored under `migrations/`
- ✅ Central configuration in `config/config.yaml`

**Use when**:

- You want to learn how to build **non-HTTP / infra-style services** with Lokstra.  
- You need an example of **syncing configuration/state** using Postgres and Lokstra’s helpers.  
- You want to see how to wire **db pools, sync map, migrations, and annotations** together.

---

### 3. [`03_tenant_management`](./03_tenant_management/)

**Tenant service with Postgres-backed store**

This template focuses on a **single bounded context (tenant management)** and shows:

- ✅ Annotation-based service & router for `tenant-service`
- ✅ Postgres-backed tenant store configured via `config/config.yaml`
- ✅ Usage of built-in middleware (`recovery`, `request_logger`)
- ✅ Simple but realistic domain + repository pattern

**Use when**:

- You want a **focused example** of a single service built with Track 2.  
- You need a reference for **tenant management / multi-tenant style** building blocks.  
- You want to see a smaller example than `01_enterprise_router_service` but still using DB.

---

## 🎯 Which Template Should I Start With?

| Your Situation | Recommended Template |
|---------------|----------------------|
| Learning enterprise Track 2 end-to-end | **01_enterprise_router_service** |
| Want infra/background service example | **02_sync_config** |
| Need focused tenant management example | **03_tenant_management** |
| Evaluating annotations + generated routers | **01_enterprise_router_service** |
| Evaluating DB integration & migrations | **02_sync_config** or **03_tenant_management** |

If you are **new to Lokstra**, first complete the router templates in
`project_templates/01_router`, then move to `01_enterprise_router_service`.

---

## 🚀 Quick Start

> Run these commands from the **repository root** (where `go.mod` is).

### 1. Enterprise Router Service

```bash
go run ./project_templates/02_app_framework/01_enterprise_router_service
```

Server starts on `http://localhost:3000` (see `config/deployment.yaml`).

- Open `01_enterprise_router_service/test.http` and `test-microservice.http` in VS Code.  
- Click **“Send Request”** to try the monolith and microservice flows.

---

### 2. Sync Config Service

```bash
go run ./project_templates/02_app_framework/02_sync_config
```

The service will:

- Initialize Postgres using migrations in `migrations/`.  
- Use `config/config.yaml` for db pool & deployment settings.

Use `02_sync_config/test.http` (if present) or your own HTTP client/tooling
to interact with the service as needed.

---

### 3. Tenant Management Service

```bash
go run ./project_templates/02_app_framework/03_tenant_management
```

Server starts on `http://localhost:3000` (see `config/config.yaml`).

- Open `03_tenant_management/tenant-service.http` in VS Code.  
- Click **“Send Request”** to test the tenant APIs.

---

## 🎓 Learning Path (Track 2 – Application Framework)

1. **Router basics** (Track 1):
   - `project_templates/01_router/01_router_only/`
   - `project_templates/01_router/02_single_app/`
   - `project_templates/01_router/03_multi_app/`
2. **Enterprise framework example**:
   - `project_templates/02_app_framework/01_enterprise_router_service/`
3. **Infrastructure / background service**:
   - `project_templates/02_app_framework/02_sync_config/`
4. **Domain-focused service (tenant)**:
   - `project_templates/02_app_framework/03_tenant_management/`
5. **Deep dive**:
   - Read `docs/02-framework-guide/` on the main docs site  
     (`https://primadi.github.io/lokstra/`).

---

## 🛠 Prerequisites

- **Go 1.23+**
- **PostgreSQL** for templates that use db pools (see each template’s `config.yaml`).
- **VS Code** (recommended) with REST Client extension for `.http` files.

---

## 📞 Support

- **Documentation**: [Lokstra Docs](https://primadi.github.io/lokstra/)
- **Issues**: [GitHub Issues](https://github.com/primadi/lokstra/issues)
- **Examples**: See other templates in `project_templates/`

---

## 📄 License

These templates are part of the Lokstra framework. See LICENSE file in project root.

---

## 🎉 Get Started

1. **Pick a template** based on your use case.  
2. **Read the template’s README** (inside its folder).  
3. **Run the example**, explore `test.http`/`.http` files.  
4. **Adapt the patterns** to your own domain and infrastructure.

Happy coding with Lokstra! 🚀
