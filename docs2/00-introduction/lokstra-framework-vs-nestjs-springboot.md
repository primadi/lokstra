---
layout: docs
title: Lokstra Framework vs NestJS / Spring Boot
description: How Lokstra’s application framework compares to popular full-stack frameworks.
---

## Same Mental Model

In **Track 2 (Application Framework)**, Lokstra targets a similar space as:

- **NestJS** (Node)  
- **Spring Boot** (Java)  
- **Uber Fx / Wire** (Go DI, partially)

You think in terms of:

- Services and modules  
- Dependency injection  
- Configuration and environments  
- Deployment topologies (monolith vs microservices)

## What’s Similar

- **Services + DI**  
  - Register services and their dependencies.  
  - Inject what you need into constructors/fields.

- **Config‑driven behavior**  
  - YAML files describe services, db pools, deployments, servers.  
  - Same code can run as monolith or microservices.

- **Annotation‑based routing**  
  - `@RouterService`, `@Route`, `@Inject` play a role similar to decorators/annotations.

## What’s Different (Go‑style)

- **Type‑safe generics, no reflection in the hot path**  
  - Lazy DI uses generics instead of `any`/interface hacks.

- **Opt‑in complexity**  
  - You can stay in router mode.  
  - You can use DI + YAML only where it helps.

- **Deployment‑first design**  
  - YAML describes **where** services run.  
  - Lokstra decides whether a dependency is local or remote and can proxy automatically.

If you like the productivity of NestJS/Spring Boot but want **minimal magic
and plain Go code**, Lokstra’s framework track is designed for that.


