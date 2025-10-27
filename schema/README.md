# Lokstra Configuration Schema

This directory contains the JSON Schema for Lokstra configuration files.

## Usage

Add this line at the top of your `config.yaml`:

```yaml
# yaml-language-server: $schema=https://primadi.github.io/lokstra/schema/lokstra.schema.json
```

## Benefits

- ✅ **Autocomplete** - VS Code suggests valid configuration keys
- ✅ **Validation** - Catches errors before runtime
- ✅ **Documentation** - Inline descriptions for all options
- ✅ **Type Safety** - Ensures correct value types

## Requirements

Install the YAML extension in VS Code:
- **Extension**: [YAML by Red Hat](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)

## Schema URL

**Permanent URL:**
```
https://primadi.github.io/lokstra/schema/lokstra.schema.json
```

This URL is served via GitHub Pages and will remain stable across versions.

## Source

The source schema is maintained at:
```
core/deploy/schema/lokstra.schema.json
```

This copy in `docs/schema/` is published to GitHub Pages for public access.
