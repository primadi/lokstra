# Lokstra Framework – Positioning Statement

## Purpose

Lokstra is a backend framework for Go developers who want a **structured, scalable, and developer-friendly way to build web services and APIs**. It is designed to reduce boilerplate, enforce clear conventions, and support both monolith and microservice architectures.

Lokstra is not a platform, not a service mesh, and not a low-level network proxy. It is a lightweight and practical framework that enhances the Go ecosystem without hiding its core principles.

---

## What Lokstra **Is**

| Domain                | Lokstra Provides                                               |
| --------------------- | -------------------------------------------------------------- |
| Backend structure     | Modular project layout, app/server separation                  |
| Routing               | Fast and flexible routing engine with middleware at all levels |
| Service registration  | Plug-and-play service modules with config and lifecycle hooks  |
| Configuration         | YAML + ENV override support                                    |
| Multi-tenancy         | Built-in tenant-aware service structure                        |
| Deployment            | Monolith, multi-binary, and microservice friendly              |
| Developer Experience  | Binding helpers, response helpers, context-aware middleware    |
| Observability (basic) | Prometheus metrics, structured logger                          |

---

## What Lokstra **Is Not**

| Not a...              | Explanation                                                    |
| --------------------- | -------------------------------------------------------------- |
| Service mesh          | Does not handle traffic routing between services (e.g., Istio) |
| APM or Tracing system | Does not inject distributed tracing (yet)                      |
| Operator framework    | No K8s controller/operator support at runtime (yet)            |
| RPC framework         | Not a gRPC or internal RPC tool (planned as `lokstra-call`)    |
| Web framework with UI | Lokstra is backend-only; frontend handled separately           |
| JVM / GraalVM rival   | Lokstra targets native Go development, not Java replacement    |

---

## Use Case Fit

Lokstra is suitable for:

* Go developers building APIs, internal services, or microservices
* Teams that want convention and structure without excessive abstraction
* Lightweight deployments (Docker, VPS, Kubernetes)
* Multi-tenant applications
* Applications that grow over time from monolith → service split

Lokstra is **not ideal** for:

* Complex streaming workloads
* Graph-heavy systems needing auto schema
* Environments needing dynamic hot reload or live instrumentation (yet)

---

## Summary

> Lokstra is the Rails-inspired framework for Go backend services. It aims to combine the simplicity of Go with the structure of a real framework — without sacrificing control.

Tagline: **Simple. Scalable. Structured.**
