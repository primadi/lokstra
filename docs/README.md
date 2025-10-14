# Lokstra Framework Documentation

> **Build Go REST APIs with less boilerplate, more flexibility**

Welcome to Lokstra! This documentation will guide you from your first "Hello World" to building production-ready microservices.

---

## 🚀 Quick Links

- **New to Lokstra?** → Start with [Introduction](00-introduction/README.md)
- **Want to code now?** → Jump to [Quick Start](00-introduction/quick-start.md)
- **Building your first API?** → Follow [Essentials](01-essentials/README.md)
- **Need advanced features?** → Explore [Deep Dive](02-deep-dive/README.md)
- **API Reference?** → Check [API Reference](03-api-reference/README.md)

---

## 📚 Documentation Structure

### [00 - Introduction](00-introduction/)
**Time**: 15 minutes • **Level**: Beginner

Understand what Lokstra is, why it exists, and what problems it solves.

- [What is Lokstra?](00-introduction/README.md) - Overview and philosophy
- [Why Lokstra?](00-introduction/why-lokstra.md) - Problems & solutions
- [Architecture](00-introduction/architecture.md) - High-level design
- [Key Features](00-introduction/key-features.md) - What makes Lokstra special
- [Quick Start](00-introduction/quick-start.md) - Your first Lokstra app in 5 minutes

---

### [01 - Essentials](01-essentials/)
**Time**: 2-3 hours • **Level**: Beginner

Learn the core concepts and build working applications. After this section, you'll be able to create production-ready REST APIs.

#### Core Components:
1. **[Router](01-essentials/01-router/)** - HTTP routing and handlers
2. **[Service](01-essentials/02-service/)** - Business logic organization
3. **[Middleware](01-essentials/03-middleware/)** - Request/response processing
4. **[Configuration](01-essentials/04-configuration/)** - App configuration patterns
5. **[App & Server](01-essentials/05-app-and-server/)** - Application lifecycle
6. **[Putting It Together](01-essentials/06-putting-it-together/)** - Complete working example

**Each section includes:**
- 📖 Concepts explained simply
- 💡 Common use cases
- 🔧 Runnable examples
- ✅ Best practices

---

### [02 - Deep Dive](02-deep-dive/)
**Time**: 4-6 hours • **Level**: Intermediate to Advanced

Master Lokstra's advanced features and internal mechanisms.

1. **[Router](02-deep-dive/router/)** - All handler forms, lifecycle, advanced patterns
2. **[Service](02-deep-dive/service/)** - DI, remote services, layered architecture
3. **[Middleware](02-deep-dive/middleware/)** - Custom middleware, advanced patterns
4. **[Configuration](02-deep-dive/configuration/)** - Multi-deployment, advanced strategies
5. **[App & Server](02-deep-dive/app-and-server/)** - Lifecycle hooks, multiple servers

---

### [03 - API Reference](03-api-reference/)
**Type**: Reference • **Level**: All levels

Complete API documentation for all packages and interfaces.

- Core packages (router, service, middleware, config, app, server)
- Helper packages (request, response, validation)
- Registry system
- Client libraries

---

### [04 - Guides](04-guides/)
**Type**: How-To Guides • **Level**: Intermediate

Practical guides for specific use cases and patterns.

- Authentication & Authorization
- Database Integration
- Testing Strategies
- Deployment Patterns
- Performance Optimization
- Migration from Other Frameworks

---

### [05 - Examples](05-examples/)
**Type**: Complete Applications • **Level**: All levels

Full working applications demonstrating real-world patterns.

- Blog API (CRUD, auth, pagination)
- E-commerce Backend (complex business logic)
- Microservices Architecture (distributed system)
- API Gateway Pattern
- Single Binary Multi-Deployment

---

## 🎯 Learning Paths

### Path 1: "I Want to Build APIs Fast"
**Recommended for**: New Lokstra users, pragmatic developers

1. [Quick Start](00-introduction/quick-start.md) - 5 min
2. [Router Essentials](01-essentials/01-router/) - 30 min
3. [Service Essentials](01-essentials/02-service/) - 30 min
4. [Complete Example](01-essentials/06-putting-it-together/) - 30 min
5. Start building! 🚀

**Total time**: ~2 hours to working API

---

### Path 2: "I Want to Master Lokstra"
**Recommended for**: Architects, framework enthusiasts

1. Complete [Introduction](00-introduction/) section
2. Work through all [Essentials](01-essentials/) examples
3. Study [Deep Dive](02-deep-dive/) sections
4. Build one [Complete Example](05-examples/)
5. Explore [Advanced Guides](04-guides/)

**Total time**: ~2-3 days to mastery

---

### Path 3: "I Have Specific Questions"
**Recommended for**: Experienced developers, specific use cases

1. Read [Architecture](00-introduction/architecture.md) - understand the big picture
2. Use Search or jump directly to relevant sections
3. Check [API Reference](03-api-reference/) for specific APIs
4. Browse [Guides](04-guides/) for patterns

**Total time**: As needed

---

## 🧭 Navigation Tips

### Finding What You Need:

- **Concepts & Theory** → Introduction + Essentials
- **Code Examples** → Every section has `/examples` folder
- **API Details** → API Reference section
- **Real-World Patterns** → Guides + Examples
- **Troubleshooting** → Each component's deep dive section

### Running Examples:

All examples are runnable! Each example folder contains:
- `main.go` - Working code
- `README.md` - What it demonstrates
- Test files or `test.http` - How to test it

```bash
# Run any example:
cd docs/01-essentials/01-router/examples/01-basic-routes
go run main.go
```

---

## 🆘 Getting Help

### Documentation Issues:
- Found a typo? Missing info? [Open an issue](https://github.com/primadi/lokstra/issues)
- Contribute improvements via Pull Request

### Framework Issues:
- [GitHub Issues](https://github.com/primadi/lokstra/issues)
- [Discussions](https://github.com/primadi/lokstra/discussions)

### Community:
- Discord (coming soon)
- Stack Overflow tag: `lokstra`

---

## 📝 Documentation Conventions

Throughout this documentation:

- 📖 **Theory/Concept** - Explains "what" and "why"
- 💡 **Example** - Shows "how"
- ⚠️ **Important** - Pay attention!
- 💭 **Tip** - Best practice or helpful hint
- 🚫 **Don't** - Common mistake to avoid
- ✅ **Do** - Recommended approach

**Code Blocks:**
```go
// ✅ Good example - recommended pattern
func GoodExample() { }

// 🚫 Bad example - avoid this
func BadExample() { }
```

---

## 🗺️ What's Next?

### If you're brand new:
👉 Start with [What is Lokstra?](00-introduction/README.md)

### If you want to code immediately:
👉 Jump to [Quick Start](00-introduction/quick-start.md)

### If you're migrating from another framework:
👉 Read [Why Lokstra?](00-introduction/why-lokstra.md) first

---

**Happy coding with Lokstra! 🚀**
