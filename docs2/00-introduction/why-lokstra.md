---
layout: docs
title: Why Lokstra
description: Motivation and design goals behind Lokstra.
---

## Why Lokstra Exists

Lokstra was created for Go teams that need **both**:

- A **great HTTP router** (like Gin/Echo/Chi).  
- A **scalable application framework** (like NestJS/Spring Boot) – but idiomatic Go.

Instead of choosing two different libraries, Lokstra lets you:

- Start small with **router-only mode**.  
- Grow into **service‑oriented, config‑driven apps** without rewriting everything.

## Design Goals

- **Track 1 – Router**  
  - Feel as simple as Gin/Echo.  
  - Powerful handler signatures and request binding.  
  - Clean, composable middleware.

- **Track 2 – Application Framework**  
  - Type‑safe lazy dependency injection (no `any` casting).  
  - Annotation‑based routers (`@RouterService`, `@Route`, `@Inject`).  
  - YAML‑driven deployments: monolith ↔ microservices with the same code.

If you are a Go developer who likes explicit code but wants some of the
productivity of NestJS or Spring Boot, Lokstra is aimed at you.


