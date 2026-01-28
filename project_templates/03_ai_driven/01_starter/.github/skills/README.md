# Lokstra Framework - AI Agent Skills

This folder contains AI agent skills for developing applications with the Lokstra framework. These skills follow the [agentskills.io](https://agentskills.io) standard for maximum discoverability and efficiency.

## Skills Overview

### Phase 1: Design & Planning

| Skill | Purpose | Output |
|-------|---------|--------|
| [design-lokstra-overview](design-lokstra-overview/) | Framework fundamentals | Framework knowledge |
| [design-lokstra-brd-generation](design-lokstra-brd-generation/) | Business requirements | `docs/modules/{project}/BRD-*.md` |
| [design-lokstra-module-requirements](design-lokstra-module-requirements/) | Module breakdown (DDD) | `docs/modules/{module}/REQUIREMENTS-*.md` |
| [design-lokstra-api-specification](design-lokstra-api-specification/) | API endpoint specs | `docs/modules/{module}/API_SPEC.md` |
| [design-lokstra-schema-design](design-lokstra-schema-design/) | Database schema | `docs/modules/{module}/SCHEMA.md`, `migrations/` |

### Phase 2: Implementation (Coming Soon)

Micro-skills for focused, efficient development:
- `implementation-lokstra-init-framework` - Initialize main.go, lokstra.Bootstrap()
- `implementation-lokstra-yaml-config` - Create configs/ with multi-file YAML
- `implementation-lokstra-create-handler` - Create @Handler with @Route
- `implementation-lokstra-create-service` - Create @Service for infrastructure
- `implementation-lokstra-create-migrations` - Create UP/DOWN migration files
- `implementation-lokstra-generate-http-files` - Generate .http client files

### Phase 3: Advanced (Coming Soon)

- `advanced-lokstra-tests` - Unit & integration tests
- `advanced-lokstra-middleware` - Custom middleware creation
- `advanced-lokstra-validate-consistency` - Validation & consistency checks

## Development Workflow

```
1. design-lokstra-overview → Understand framework
2. design-lokstra-brd-generation → Define business requirements
3. design-lokstra-module-requirements → Break down into modules
4. design-lokstra-api-specification → Design API endpoints
5. design-lokstra-schema-design → Design database schema
   ↓ (Phase 2 - Implementation skills coming)
6. implementation-lokstra-* → Generate code
   ↓ (Phase 3 - Advanced skills)
7. advanced-lokstra-* → Add tests, middleware, validation
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
