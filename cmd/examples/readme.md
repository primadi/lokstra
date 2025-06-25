# Lokstra Examples Overview

This folder contains categorized, progressive examples demonstrating all features of the **Lokstra Framework**. Each directory is self-contained and executable, and also acts as a form of integration test.

All examples are written in **Go**, with a clear separation of concepts to help developers explore and adopt Lokstra step by step.

---

## 📘 1. `basic_overview/`

Introductory examples, showing how Lokstra can scale from simple router usage to full server with YAML configuration.

* `01_minimal_router_only/` – Just the router
* `02_router_with_app/` – Adds App for port and middleware
* `03_server_with_apps/` – Adds Server + multi App
* `04_with_logger_service/` – Using LoggerService
* `05_with_yaml_config/` – Using YAML-based configuration

---

## 🚦 2. `router_features/`

Advanced routing capabilities including grouping, mounting, and middleware layering.

* `01_group_and_nested_routes/`
* `02_middleware_usage/`
* `03_mount_static/`
* `04_mount_spa/`
* `05_mount_reverse_proxy/`

---

## 🧑‍🏫 3. `best_practices/`

Recommended patterns for better maintainability and developer experience.

* `01_custom_request_context/`
* `02_named_handlers/`
* `03_split_config_files/`

---

## 🧩 4. `customization/`

How to override and extend core Lokstra behavior.

* `01_custom_json_formatter/`
* `02_custom_response_wrapper/`
* `03_override_http_methods/`
* `04_custom_router_engine/`

---

## 🧱 5. `service_lifecycle/`

Covers service registration, hooks, and shutdown flows.

* `01_register_named_service/`
* `02_access_service/`
* `03_hook_on_server_start/`
* `04_shutdown_hook/`

---

## 🏢 6. `business_services/`

Examples of domain-driven, custom services containing business logic.

* `01_ledger_service/`
* `02_loan_service/`
* `03_inventory_service/`

---

## 🛠 7. `default_services/`

Demonstrates built-in Lokstra services.

* `01_logger/`
* `02_dbpool/`
* `03_redis/`
* `04_jwt_auth/`
* `05_email_sender/`
* `06_metrics/`
* `07_healthcheck/`

---

## 🧰 8. `default_middleware/`

Prebuilt middleware available in Lokstra.

* `01_recovery/`
* `02_request_logger/`
* `03_cors/`
* `04_jwt_auth_middleware/`
* `05_custom_middleware/`

---

Each directory will include a minimal `main.go` and inline comments for clarity. You can run any example directly and inspect the behavior.

---

> **Tip**: Start with `basic_overview/` and work your way down. Lokstra is modular and grows with your needs.
