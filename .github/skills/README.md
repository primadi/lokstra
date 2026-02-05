# Lokstra Framework - AI Agent Skills

This folder contains AI agent skills for developing applications with the Lokstra framework. These skills follow the [agentskills.io](https://agentskills.io) standard for maximum discoverability and efficiency.

## Skills Overview

### Phase 1: Design & Planning ✅ COMPLETE

| Skill | Purpose | Output |
|-------|---------|--------|
| [design-lokstra-overview](design-lokstra-overview/) | Framework fundamentals | Framework knowledge |
| [design-lokstra-brd-generation](design-lokstra-brd-generation/) | Business requirements | `docs/modules/{project}/BRD-*.md` |
| [design-lokstra-module-requirements](design-lokstra-module-requirements/) | Module breakdown (DDD) | `docs/modules/{module}/REQUIREMENTS-*.md` |
| [design-lokstra-api-specification](design-lokstra-api-specification/) | API endpoint specs | `docs/modules/{module}/API_SPEC.md` |
| [design-lokstra-schema-design](design-lokstra-schema-design/) | Database schema | `docs/modules/{module}/SCHEMA.md`, `migrations/` |

### Phase 2: Implementation ✅ COMPLETE

Micro-skills for focused, efficient development:

| Skill | Purpose | Output |
|-------|---------|--------|
| [implementation-lokstra-init-framework](implementation-lokstra-init-framework/) | Initialize main.go, lokstra.Bootstrap() | Working main.go with services |
| [implementation-lokstra-yaml-config](implementation-lokstra-yaml-config/) | Create configs/ with multi-file YAML | `configs/config.yaml` |
| [implementation-lokstra-create-handler](implementation-lokstra-create-handler/) | Create @Handler with @Route | HTTP endpoint handlers |
| [implementation-lokstra-create-service](implementation-lokstra-create-service/) | Create @Service for infrastructure | Repository/service implementations |
| [implementation-lokstra-create-migrations](implementation-lokstra-create-migrations/) | Create UP/DOWN migration files | `migrations/{module}/*.sql` |
| [implementation-lokstra-generate-http-files](implementation-lokstra-generate-http-files/) | Generate .http client files | `api/*.http` test files |

### Phase 3: Advanced (3 skills) ✅ COMPLETE

Advanced features for production-ready applications:

| Skill | Purpose | Output |
|-------|---------|--------|
| [advanced-lokstra-tests](advanced-lokstra-tests/) | Unit & integration tests, mocks | Test files, coverage reports |
| [advanced-lokstra-middleware](advanced-lokstra-middleware/) | Custom middleware, auth, logging | Middleware implementations |
| [advanced-lokstra-validate-consistency](advanced-lokstra-validate-consistency/) | Dependency validation, config checks | Validation scripts, CI/CD integration |

## Development Workflow

```
1. design-lokstra-overview → Understand framework
2. design-lokstra-brd-generation → Define business requirements
3. design-lokstra-module-requirements → Break down into modules
4. design-lokstra-api-specification → Design API endpoints
5. design-lokstra-schema-design → Design database schema
   ↓ DESIGN PHASE COMPLETE
6. implementation-lokstra-init-framework → Setup main.go
7. implementation-lokstra-yaml-config → Configure services
8. implementation-lokstra-create-handler → Create endpoints
9. implementation-lokstra-create-service → Create repositories
10. implementation-lokstra-create-migrations → Create migrations
11. implementation-lokstra-generate-http-files → Create test files
    ↓ IMPLEMENTATION PHASE COMPLETE
12. advanced-lokstra-tests → Write unit & integration tests
13. advanced-lokstra-middleware → Create custom middleware
14. advanced-lokstra-validate-consistency → Validate & deploy
```

## Skill Format (agentskills.io compliant)

Each skill follows this structure:

```
{phase}-{skill-name}/
├── SKILL.md          # Main skill instructions with YAML frontmatter
├── references/       # Detailed templates and examples (optional)
├── scripts/          # Helper scripts (optional)
└── assets/           # Images, diagrams (optional)
```

### YAML Frontmatter

```yaml
---
name: skill-name
description: Clear 1-1024 char description of when to use this skill
license: MIT
metadata:
  author: lokstra-framework
  version: "1.0"
  framework: lokstra
  phase: design | implementation | advanced
  order: 1
compatibility: Designed for GitHub Copilot, Cursor, Claude Code
---
```

## Using These Skills

### With GitHub Copilot

Copilot automatically reads `.github/skills/` and provides context-aware suggestions.

### With Cursor / Claude Code

Reference specific skills:

```
@design-lokstra-brd-generation Generate BRD for user authentication
@implementation-lokstra-create-handler Create user handler with routes
```

### Manual Reference

Read `SKILL.md` files directly for step-by-step instructions.

## Updating Skills

When Lokstra framework is updated:

```bash
lokstra update-skills
```

This updates skills in your project's `.github/skills/` folder.

## Creating Custom Skills

You can add custom project-specific skills:

1. Create folder: `.github/skills/{phase}-{custom-skill}/`
2. Create `SKILL.md` with YAML frontmatter
3. Add references/templates as needed

## Skill Naming Convention

- **Folder:** `{phase}-{skill-name}` (e.g., `design-lokstra-overview`)
- **YAML name:** `skill-name` without phase (e.g., `lokstra-overview`)
- **Phases:** `design`, `implementation`, `advanced`

## Resources

- **Lokstra Documentation:** https://primadi.github.io/lokstra/
- **agentskills.io Standard:** https://agentskills.io
- **Skill Roadmap:** [SKILL_ROADMAP.md](SKILL_ROADMAP.md)
- **Example Projects:** https://github.com/primadi/lokstra/tree/main/examples

## License

MIT License - See [LICENSE](../../LICENSE) for details
