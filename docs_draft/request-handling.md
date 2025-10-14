# Lokstra Request Handling Flow

This document explains the request handling flow in Lokstra, designed to help developers understand how HTTP requests are processed from the listener to the final response.

## Overview

Lokstra uses a layered approach for request handling, combining router, engine, handler chains, and context management. The flow ensures middleware and error handling are robust and extensible.

## Request Flow Steps

1. **Listener to Router**
   - The HTTP listener receives an incoming request and forwards it to the router.
   - The router builds the engine if needed, then calls `engine.ServeHTTP` to process the request. (See `router_impl.go`)

2. **RouterEngine to Handler**
   - The `RouterEngine` invokes the adapted handler, which calls `request.Handler.ServeHTTP`. (See `request/handler.go`)

3. **Handler ServeHTTP: Context & Chain Execution**
   - `ServeHTTP` creates a new `Context` for the request and response.
   - It calls `c.ExecuteHandler()`, which executes all handlers and middleware in order using `c.Next()`.
   - After all handlers are executed, `c.FinalizeResponse()` is called to determine the final response.

4. **FinalizeResponse: Response Priority**
   - If `W.ManuallyWritten()` is true (response already written manually), do nothing further.
   - If any handler returns a non-nil error, read `RespStatusCode`.
   - If the status code is still 0 or less than `http.StatusBadRequest`, set it to Internal Server Error (500) and return an error response.
   - Otherwise, write the response with the correct status code and body.

## Key Principles

- **Manual Response Priority:** If the response is written manually, Lokstra will not override it.
- **Error Handling:** Errors returned by handlers are prioritized and mapped to appropriate HTTP status codes.
- **Middleware Chain:** Handlers and middleware are executed in order, with each calling `c.Next()` to continue the chain.
- **Robust Finalization:** The final response is determined by error status, manual writes, and status code checks to prevent double responses or inconsistent status.

## Example Sequence

1. Request received by listener
2. Router builds engine and calls `engine.ServeHTTP`
3. Engine calls `request.Handler.ServeHTTP`
4. Handler creates context, executes handler chain
5. FinalizeResponse checks manual write, errors, and status code
6. Response sent to client

---

This flow ensures that Lokstra applications are predictable, extensible, and easy to debug for developers building HTTP APIs and web services.
