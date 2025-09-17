# 🔧 Lokstra `ServiceURI` Naming Convention

In Lokstra, services are identified using a unified string format called `ServiceURI`.  
This format standardizes service lookup across modules, YAML config, plugins, and remote service calls.

---

## 📌 Format

```
lokstra://<InterfaceName>[/<InstanceName>]
lokstra://<PackageName>.<InterfaceName>/<InstanceName>
```

---

## 🧠 Interpretation Rules

| Component        | Description                                                                 |
|------------------|-----------------------------------------------------------------------------|
| `lokstra://`     | Required prefix to identify this as a Lokstra ServiceURI                    |
| `InterfaceName`  | CamelCase name of the service interface                                     |
| `PackageName`    | (Optional) If omitted, defaults to the standard `serviceapi` package        |
| `InstanceName`   | Required identifier for the specific instance of the service                |

---

## ✅ Examples

| URI                                      | Interface      | Package       | Instance     |
|------------------------------------------|----------------|----------------|--------------|
| `lokstra://Logger/default`               | `Logger`       | `serviceapi`   | `default`    |
| `lokstra://DbPoolPg/read`                | `DbPoolPg`     | `serviceapi`   | `read`       |
| `lokstra://rpc_service.RpcService/hello` | `RpcService`   | `rpc_service`  | `hello`      |
| `lokstra://my_module.MyService/dev`      | `MyService`    | `my_module`    | `dev`        |

---

## 🛠 Validation Rules

- Must begin with `lokstra://`
- Interface name must be in **CamelCase**
- Instance name is **required**
- Package name must be a valid Go identifier (if provided)
- Use `.` to separate package and interface name

---

## 🧩 Notes

- Used for service registration, resolution, and configuration
- Compatible with both internal and external service interfaces
- The actual Go interface must be implemented manually; Go does **not** resolve interface by string at runtime

---

## 🔍 Examples in Go Code

```go
svc := ctx.GetService("lokstra://Logger/default").(serviceapi.Logger)

svc := ctx.GetService("lokstra://my_package.MyService/dev").(my_package.MyService)
```

---

## 🧪 Tips for Developers

- Use short instance names like `default`, `read`, `write`, or environment-specific like `dev`, `prod`
- Prefer `CamelCase` for interface names and `snake_case` for package names

---