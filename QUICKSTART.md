# Lokstra CLI - Quick Start Guide

## Installation

Install Lokstra CLI globally:

```bash
go install github.com/primadi/lokstra/cmd/lokstra@latest
```

Verify installation:

```bash
lokstra version
```

## Create Your First Project

### Interactive Mode (Recommended)

```bash
lokstra new my-api
```

This will show you an interactive menu to select from available templates.

### Quick Start with Specific Template

```bash
# Simple router-based API
lokstra new my-api -template 01_router/01_router_only

# Medium-sized application with DDD
lokstra new blog-api -template 02_app_framework/01_medium_system

# Enterprise application with bounded contexts
lokstra new enterprise-app -template 02_app_framework/03_enterprise_router_service
```

### Run Your New Project

```bash
cd my-api
go run .
```

That's it! Your Lokstra application is now running.

### Generate Code (for Enterprise Templates)

If you're using Enterprise Router Service template with annotations:

```bash
# After modifying annotations in your code
lokstra autogen

# Or specify the project folder
lokstra autogen ./my-api
```

This regenerates routers based on your annotations.

## What Templates Are Available?

Run the CLI in interactive mode to see all available templates:

```bash
lokstra new myapp
```

### Router Patterns (Learning & Simple APIs)

1. **Router Only** - Pure routing, no framework overhead
2. **Single App** - Production-ready single application
3. **Multi App** - Multiple apps in one server

### Framework Patterns (Production Applications)

4. **Medium System** - Domain-driven design for medium apps
5. **Enterprise Modular** - DDD with bounded contexts
6. **Enterprise Router Service** - Annotation-based enterprise

## What the CLI Does Automatically

✅ Downloads template from GitHub  
✅ Copies all files to your project  
✅ Fixes all import paths automatically  
✅ Runs `go mod init <your-project-name>`  
✅ Runs `go mod tidy` to fetch dependencies  

**No manual configuration needed!**

## Advanced Usage

### Use Different Branch

```bash
lokstra new myapp -template 01_router/01_router_only -branch main
```

### Get Help

```bash
lokstra help
```

## Next Steps

After creating your project:

1. Read the project's README: `cat README.md`
2. Check the test.http file for API examples
3. Explore the code structure
4. Modify for your needs
5. Deploy!

## Documentation

- Full CLI Documentation: [cmd/lokstra/README.md](./cmd/lokstra/README.md)
- Template Documentation: [project_templates/README.md](./project_templates/README.md)
- Framework Documentation: [https://primadi.github.io/lokstra/](https://primadi.github.io/lokstra/)

## Support

- Issues: [GitHub Issues](https://github.com/primadi/lokstra/issues)
- Discussions: [GitHub Discussions](https://github.com/primadi/lokstra/discussions)

---

**Happy coding with Lokstra! 🚀**
