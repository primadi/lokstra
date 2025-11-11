---
layout: docs
title: Examples
---

# Lokstra Examples

> 🎯 **Two learning tracks: Router-only or Full Framework**

Choose your learning path based on how you want to use Lokstra.

---

## 🎯 Choose Your Track

### Track 1: Router Only (Like Echo, Gin, Chi)
**Time**: 2-3 hours • **Use Lokstra as a flexible HTTP router**

Learn routing, handlers, and middleware without dependency injection complexity.

👉 **[Start with Router Examples](./router-only/)**

**What you'll learn:**
- ✅ Basic routing and handlers
- ✅ 29 handler form variations
- ✅ Middleware patterns (global, per-route, groups)
- ✅ Quick prototyping

**Perfect for:**
- Quick APIs and prototypes
- Learning HTTP routing fundamentals
- Developers familiar with Echo, Gin, or Chi
- Projects that don't need DI

---

### Track 2: Full Framework (Like NestJS, Spring Boot)
**Time**: 8-12 hours • **Use Lokstra as a complete application framework**

Learn services, dependency injection, auto-routers, annotations, and deployment patterns.

👉 **[Start with Framework Examples](./full-framework/)**

**What you'll learn:**
- ✅ Service layer and dependency injection
- ✅ **Annotation-driven development** (`@RouterService`, `@Inject`, `@Route`)
- ✅ Auto-generated REST routers from service methods
- ✅ Configuration-driven deployment (YAML or Code)
- ✅ Monolith → Microservices migrations
- ✅ External service integration

**Perfect for:**
- Enterprise applications
- Microservices architectures
- Teams wanting DI and auto-router
- Production-scale projects
- **Developers familiar with NestJS decorators or Spring annotations**

---

## 📚 Complete Feature Map

| Feature | Track 1 (Router) | Track 2 (Framework) |
|---------|------------------|---------------------|
| **HTTP Routing** | ✅ Core focus | ✅ Included |
| **Handler Forms** | ✅ 29 variations | ✅ Same flexibility |
| **Middleware** | ✅ Global, per-route | ✅ Plus registry-based |
| **Services** | ❌ Not covered | ✅ Core pattern |
| **Dependency Injection** | ❌ Not needed | ✅ Lazy, type-safe |
| **Annotations** | ❌ Not covered | ✅ `@RouterService`, `@Inject`, `@Route` |
| **Auto-Router** | ❌ Manual only | ✅ From services |
| **Configuration** | ❌ Code only | ✅ YAML or Code |
| **Microservices** | ❌ Not covered | ✅ Multi-deployment |

---

## 🔄 Can I Switch Tracks?

**Yes! Start with Track 1, upgrade to Track 2 later.**

Track 1 code is compatible with Track 2. You can:
1. Start with router-only examples (simple, fast)
2. Add services and DI when needed (gradual)
3. Enable auto-router for new features (optional)
4. Keep manual routing for existing routes (backward compatible)

**Track 1 → Track 2 is an upgrade, not a rewrite!**

---

## 🚀 Quick Start

### For Router Track:
```bash
cd docs/00-introduction/examples/router-only/01-hello-world
go run main.go
curl http://localhost:3000/
```

### For Framework Track:
```bash
cd docs/00-introduction/examples/full-framework/01-crud-api
go run main.go
curl http://localhost:3000/users
```

---

## 📖 What's Next?

After examples, continue learning:

### Completed Track 1 (Router)?
- **[Router Guide](../../01-router-guide/)** - Deep dive into routing
- **[API Reference](../../03-api-reference/)** - Complete API docs

**Want more?** → Explore Track 2 for DI and auto-router!

### Completed Track 2 (Framework)?
- **[Framework Guide](../../02-framework-guide/)** - Advanced DI patterns
- **[Configuration Schema](../../03-api-reference/03-configuration/)** - Full YAML reference
- **[Production Patterns](../../02-framework-guide/)** - Microservices deployment

---

## 💡 Comparison with Other Frameworks

### Track 1 (Router) compares with:
- **Echo** - Similar flexibility, more handler forms
- **Gin** - Similar performance, cleaner API
- **Chi** - Similar routing, more middleware options
- **Fiber** - Similar speed, Go-idiomatic (no fasthttp)

### Track 2 (Framework) compares with:
- **NestJS** (Node.js) - Similar DI and auto-router concepts
- **Spring Boot** (Java) - Similar enterprise patterns
- **Uber Fx** (Go) - Similar DI, plus auto-router
- **Buffalo** (Go) - Similar full-stack, more flexible

---

**Ready?** Choose your track:

<div style=\"display: grid; grid-template-columns: 1fr 1fr; gap: 2rem; margin: 2rem 0;\">
  <div style=\"padding: 2rem; border: 2px solid #4a9eff; border-radius: 8px; background: #1a1a2e;\">
    <h3>🎯 Track 1: Router Only</h3>
    <p><strong>Time:</strong> 2-3 hours</p>
    <p><strong>Like:</strong> Echo, Gin, Chi</p>
    <p><strong>Focus:</strong> HTTP routing</p>
    <a href=\"./router-only/\" style=\"display: inline-block; margin-top: 1rem; padding: 0.5rem 1rem; background: #4a9eff; color: white; text-decoration: none; border-radius: 4px;\">Start Router Track →</a>
  </div>
  
  <div style=\"padding: 2rem; border: 2px solid #ff6b6b; border-radius: 8px; background: #1a1a2e;\">
    <h3>🏗️ Track 2: Full Framework</h3>
    <p><strong>Time:</strong> 8-12 hours</p>
    <p><strong>Like:</strong> NestJS, Spring Boot</p>
    <p><strong>Focus:</strong> Enterprise apps</p>
    <a href=\"./full-framework/\" style=\"display: inline-block; margin-top: 1rem; padding: 0.5rem 1rem; background: #ff6b6b; color: white; text-decoration: none; border-radius: 4px;\">Start Framework Track →</a>
  </div>
</div>
