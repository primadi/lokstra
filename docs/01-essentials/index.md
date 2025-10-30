---
layout: docs
title: Essentials - Getting Started
---

# Essentials - Build Your First Lokstra API

> **Learn the fundamentals and build production-ready APIs**  
> **Time**: 2-3 hours • **Level**: Beginner • **Prerequisites**: Basic Go knowledge

---

## 🎯 What You'll Learn

After completing this section, you'll be able to:
- ✅ Create routers and register routes with multiple handler styles
- ✅ Organize business logic into services
- ✅ Apply middleware for cross-cutting concerns
- ✅ Configure applications via YAML
- ✅ Start and manage servers with graceful shutdown
- ✅ Build complete REST APIs ready for production

---

## 📚 Learning Path

Work through these sections **in order**. Each builds on the previous:

### [01 - Router](01-router/)
**Time**: 30-40 minutes • **Concepts**: 5 • **Examples**: 5

Learn HTTP routing, handler registration, and route organization.

**What you'll build**:
- Basic REST endpoints
- Routes with parameters
- Grouped routes (versioning)
- Routes with middleware

**Key Takeaways**:
- 4 essential handler forms (out of 29 total!)
- Route groups for API versioning
- Per-route vs global middleware

👉 [Start with Router](01-router)

---

### [02 - Service](02-service/)
**Time**: 40-50 minutes • **Concepts**: 6 • **Examples**: 4

Learn service patterns, dependency injection, and service-as-router.

**What you'll build**:
- Reusable service components
- Services in handlers
- Auto-generated HTTP routes from services ⭐

**Key Takeaways**:
- Service factory pattern
- Dependency injection
- **Service methods → HTTP endpoints automatically!**

👉 [Continue to Service](02-service)

---

### [03 - Middleware](03-middleware/)
**Time**: 25-30 minutes • **Concepts**: 4 • **Examples**: 3

Learn request/response processing, middleware patterns, and built-in middleware.

**What you'll build**:
- Logging middleware
- Authentication middleware
- CORS configuration

**Key Takeaways**:
- Middleware chain execution
- Global vs per-route middleware
- Using built-in middleware

👉 [Continue to Middleware](03-middleware)

---

### [04 - Configuration](04-configuration/)
**Time**: 30-35 minutes • **Concepts**: 4 • **Examples**: 3

Learn YAML configuration, environment variables, and configuration strategies.

**What you'll build**:
- YAML-based configuration
- Environment-specific configs
- Config validation

**Key Takeaways**:
- Code + Config pattern (recommended!)
- Environment variables in YAML
- Multi-file configuration

👉 [Continue to Configuration](04-configuration)

---

### [05 - App & Server](05-app-and-server/)
**Time**: 20-25 minutes • **Concepts**: 3 • **Examples**: 2

Learn application lifecycle, server management, and graceful shutdown.

**What you'll build**:
- Basic app with multiple routers
- Server with graceful shutdown

**Key Takeaways**:
- App combines routers
- Server manages apps
- Automatic graceful shutdown

👉 [Continue to App & Server](05-app-and-server)

---

### [06 - Putting It Together](06-putting-it-together/)
**Time**: 45-60 minutes • **Concepts**: Integration • **Examples**: 1 complete app

Build a complete REST API using everything you've learned.

**What you'll build**:
- Complete Todo API with:
  - CRUD operations
  - Authentication
  - Validation
  - Error handling
  - Configuration
  - Tests

**Key Takeaways**:
- How components work together
- Project structure best practices
- Production-ready patterns

👉 [Final Project](06-putting-it-together)

---

## 🎓 Teaching Philosophy

This section follows these principles:

### 1. **Progressive Complexity**
Start simple, add complexity gradually. No overwhelming information dumps.

### 2. **Runnable Examples**
Every concept has a working example you can run and modify.

### 3. **Practical First**
Learn by doing. Theory comes after you've seen it work.

### 4. **Common Patterns**
Focus on what you'll use 80% of the time, not edge cases.

### 5. **Best Practices**
Learn the right way from the start, avoiding common pitfalls.

---

## 📖 How to Use This Section

### Recommended Approach (2-3 hours):
```
Day 1 Morning: 01-Router, 02-Service     (1.5 hours)
Day 1 Afternoon: 03-Middleware, 04-Config (1 hour)
Day 2 Morning: 05-App & Server            (30 min)
Day 2 Afternoon: 06-Complete Example      (1 hour)
```

### Fast Track (1 hour):
```
1. Read each index
2. Run one example per section
3. Build the final project
```

### Deep Learning (4-6 hours):
```
1. Read all content
2. Run ALL examples
3. Modify examples
4. Build variations of final project
```

---

## 🧪 Running Examples

Every example is **self-contained and runnable**.

### To run an example:
```bash
# Navigate to example folder
cd docs/01-essentials/01-router/examples/01-basic-routes

# Run it
go run main.go

# In another terminal, test it
curl http://localhost:3000/ping
```

### Example structure:
```
01-basic-routes/
├── main.go          # Working code
├── index        # What it demonstrates
└── test.http        # Test requests (optional)
```

---

## 💡 Learning Tips

### For Each Section:

1. **Read** the concept explanation
2. **Study** the code examples in README
3. **Run** the example applications
4. **Experiment** - modify and see what happens
5. **Build** something small on your own

### Don't Skip Ahead!
Each section builds on previous knowledge. Skipping will cause confusion.

### Stuck? 
- Re-read the section
- Check the complete example in section 06
- Look at [Deep Dive](../02-deep-dive) for more details
- Ask in [Discussions](https://github.com/primadi/lokstra/discussions)

---

## 🎯 After Essentials, You'll Know:

### Confident With:
- ✅ Creating REST APIs
- ✅ Organizing code with services
- ✅ Using middleware effectively
- ✅ Configuring via YAML
- ✅ Building production-ready apps

### Ready For:
- 🚀 Building real applications
- 🚀 Exploring [Deep Dive](../02-deep-dive) for advanced features
- 🚀 Reading [Complete Examples](../05-examples)
- 🚀 Implementing [Specific Patterns](../04-guides)

### Next Steps:
1. **Build Something!** - Best way to solidify learning
2. **Explore [Deep Dive](../02-deep-dive)** - Learn advanced patterns
3. **Study [Complete Examples](../05-examples)** - Real-world applications

---

## 🗺️ Quick Reference

| Component | What It Does | When to Use |
|-----------|--------------|-------------|
| **Router** | Match HTTP requests to handlers | Every API needs this |
| **Service** | Encapsulate business logic | Organize code, reuse logic |
| **Middleware** | Process requests/responses | Logging, auth, CORS, etc |
| **Configuration** | Externalize settings | Different environments |
| **App** | Combine routers | Group related features |
| **Server** | Manage apps | Multiple services/ports |

---

## 📊 Progress Tracker

Track your progress through Essentials:

- [ ] 01 - Router completed
- [ ] 02 - Service completed
- [ ] 03 - Middleware completed
- [ ] 04 - Configuration completed
- [ ] 05 - App & Server completed
- [ ] 06 - Complete Example built

**Estimated completion**: 2-3 hours of focused learning

---

**Ready to start?** 👉 [Begin with Router](01-router)
