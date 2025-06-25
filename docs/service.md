# Lokstra Service System

Lokstra Service is a flexible and efficient system for registering, managing, and interacting with internal services in the Lokstra Framework. It enables developers to build reusable modules that integrate deeply into the application lifecycle, while remaining loosely coupled.

---

## ✨ Overview

Each Lokstra Service is:

- **Named**: Every service instance is uniquely identified by its `instanceName`.
- **Configurable**: Services can receive configuration via YAML or code.
- **Optional**: Services can be enabled or disabled via configuration.
- **Extensible**: Services can hook into lifecycle events like app start or HTTP response.

---

## 📦 Minimal Interface

All services must implement the following interface:

```go
type Service interface {
	InstanceName() string        // Unique instance name
	GetConfig(key string) any    // Access configuration by key
}
