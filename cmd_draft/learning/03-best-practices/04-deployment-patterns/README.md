# 04-Deployment Patterns

Complete comparison of all three deployment strategies.

## Contents

- **comparison.md** - Detailed comparison with pros/cons
- Decision matrix
- Migration path
- Configuration examples

## Quick Reference

| Pattern | Processes | Ports | Complexity | Best For |
|---------|-----------|-------|------------|----------|
| Monolith Single | 1 | 1 | ⭐ Simple | Startups |
| Monolith Multi | 1 | 2+ | ⭐⭐ Moderate | Growing apps |
| Microservices | 3+ | 1 per service | ⭐⭐⭐ Complex | Enterprises |

## Key Insight

**The same application code works for all three patterns - only `config.yaml` changes!**

This is the power of Lokstra's config-driven architecture.

## See

- `comparison.md` for complete analysis
