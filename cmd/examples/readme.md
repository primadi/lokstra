# Lokstra Examples - Progressive Learning Path

This directory contains comprehensive, progressive examples that follow the learning path outlined in the [Lokstra documentation](../../docs/README.md). Each example is self-contained, runnable, and builds upon concepts from previous examples.

> 🎯 **Learning Objective**: Master Lokstra framework step-by-step, from basic concepts to production-ready applications.

---

## 🚀 Learning Path Overview

The examples are organized to match the documentation structure and provide hands-on experience with Lokstra's key features:

1. **Getting Started** → Foundation and basic concepts
2. **Core Features** → Smart binding, HTMX, services, middleware  
3. **Advanced Usage** → Configuration, customization, production patterns
4. **Real-World Examples** → Complete applications and best practices

---

## 📘 01. Getting Started (`01_getting_started/`)

Master the fundamentals of Lokstra applications based on [Getting Started Guide](../../docs/getting-started.md).

* `01_minimal_app/` – Simplest possible Lokstra app
* `02_smart_binding/` – Request binding with struct tags  
* `03_structured_responses/` – Response helpers and method chaining
* `04_multiple_apps/` – Server with multiple applications
* `05_graceful_shutdown/` – Production-ready lifecycle management

**Key Learning**: App creation, smart binding, structured responses, graceful shutdown.

---

## 🏗️ 02. Core Concepts (`02_core_concepts/`)

Deep dive into Lokstra's architecture based on [Core Concepts](../../docs/core-concepts.md).

* `01_registration_context/` – Dependency injection basics
* `02_request_context/` – Working with request context
* `03_type_safe_services/` – Service container and type safety
* `04_handler_patterns/` – Different handler approaches
* `05_error_handling/` – Proper error handling patterns

**Key Learning**: Registration context, request context, services, handler patterns.

---

## �️ 03. Routing & Middleware (`03_routing/`)

Advanced routing and middleware based on [Routing](../../docs/routing.md) and [Middleware](../../docs/middleware.md).

* `01_basic_routing/` – HTTP methods and path parameters
* `02_route_groups/` – Route grouping and prefixes
* `03_middleware_chain/` – Middleware pipeline and priorities
* `04_static_files/` – Static file serving and SPA support
* `05_custom_middleware/` – Creating custom middleware

**Key Learning**: Route organization, middleware pipeline, static serving.

---

## � 04. HTMX Integration (`04_htmx/`)

Modern web applications with HTMX based on [HTMX Integration](../../docs/htmx-integration.md).

* `01_basic_htmx/` – HTMX page serving
* `02_dynamic_content/` – HTMX with dynamic data
* `03_forms_and_interactions/` – Interactive forms
* `04_real_time_updates/` – Live content updates
* `05_complete_webapp/` – Full HTMX application

**Key Learning**: HTMX integration, dynamic content, interactive UIs.

---

## ⚙️ 05. Services & Configuration (`05_services/`)

Service management and configuration based on [Services](../../docs/services.md) and [Configuration](../../docs/configuration.md).

* `01_built_in_services/` – Using built-in services (Logger, DB, Redis)
* `02_custom_services/` – Creating custom services
* `03_service_factories/` – Service factories and configuration
* `04_yaml_configuration/` – YAML-based app configuration
* `05_environment_overrides/` – Environment-specific configs

**Key Learning**: Service container, configuration system, environment management.

---

## 🔧 06. Built-in Features (`06_builtin/`)

Comprehensive coverage of built-in services and middleware.

* `01_database_pool/` – PostgreSQL connection pool
* `02_redis_cache/` – Redis integration
* `03_structured_logging/` – Advanced logging patterns
* `04_metrics_monitoring/` – Prometheus metrics
* `05_health_checks/` – Application health monitoring
* `06_cors_security/` – CORS and security middleware

**Key Learning**: Production services, observability, security.

---

## � 07. Advanced Patterns (`07_advanced/`)

Advanced usage patterns and customization.

* `01_request_validation/` – Custom validation rules
* `02_response_customization/` – Custom response formats
* `03_middleware_composition/` – Advanced middleware patterns
* `04_testing_strategies/` – Testing Lokstra applications
* `05_performance_optimization/` – Performance best practices

**Key Learning**: Advanced patterns, testing, performance.

---

## 🏢 08. Real-World Examples (`08_real_world/`)

Complete applications demonstrating production patterns.

* `01_rest_api/` – Complete REST API with authentication
* `02_htmx_dashboard/` – Interactive dashboard application
* `03_microservice/` – Microservice with full observability
* `04_multi_tenant_app/` – Multi-application deployment
* `05_production_ready/` – Production deployment example

**Key Learning**: Complete applications, production patterns, deployment.

---

## 🚀 Quick Start

1. **Begin with `01_getting_started/01_minimal_app/`** to understand basics
2. **Work through each section sequentially** for comprehensive learning
3. **Run examples**: Each directory contains a runnable `main.go`
4. **Read documentation**: Examples reference specific docs sections
5. **Experiment**: Modify examples to understand behavior

### Running Examples

```bash
# Navigate to any example directory
cd 01_getting_started/01_minimal_app/

# Run the example
go run main.go

# Test with curl
curl http://localhost:8080/hello
```

### Prerequisites

- Go 1.21+ installed
- Basic understanding of Go and HTTP concepts
- Familiarity with REST APIs (helpful but not required)

---

## 📚 Documentation Integration

Each example includes:

- **📖 Clear documentation** with learning objectives
- **🔗 Links to relevant docs sections** for deeper understanding  
- **💡 Inline comments** explaining Lokstra-specific concepts
- **🧪 Test commands** to verify functionality
- **🔄 Progressive complexity** building on previous examples

---

## 🤝 Contributing

Help improve the learning experience:

- **Add examples** for new features
- **Improve documentation** and clarity
- **Fix bugs** or outdated patterns
- **Suggest improvements** to the learning path

See [Contributing Guidelines](../../CONTRIBUTING.md) for more details.

---

*Start your Lokstra journey with `01_getting_started/01_minimal_app/` and build your way up to production-ready applications!*
