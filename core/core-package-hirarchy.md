# /core Package Hierarchy

| Package      | May Import                | Description |
|--------------|--------------------------|-------------|
| `/response`  | (none)                   | Handles response formatting and types. Should remain independent and not import other core packages. |
| `/request`   | `/response`              | Contains the basic request context and types. Can import `/response` for advanced request-response handling. |
| `/service`   | `/request`               | Contains business logic and service functions. Services can use request types for input validation and processing. |
| `/middleware`| `/request`, `/service`   | Implements middleware logic, such as authentication, logging, or error handling. Middleware can access request types and call services. |
| `/router`    | `/request`, `/middleware`| Defines routing logic and maps endpoints to handlers. Routers can use request types and middleware for processing requests. |
| `/app`       | `/router`, `/service`    | The main application layer. It composes routers and services to build the application structure. |
| `/server`    | `/app`,`/service`       | The entry point for running the server. It initializes the app and may directly use services for setup or background tasks. |

---
