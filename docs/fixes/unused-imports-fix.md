# Fix: Unused Imports in Generated Code

## Problem
Generated code (`zz_generated.lokstra.go`) was including **all** import statements from source files, even if those packages were not used in handler method signatures or injected dependencies.

### Example
```go
// Source file: credential_service.go
import (
	"github.com/primadi/lokstra-auth/credential/domain"        // NOT used in handlers
	"github.com/primadi/lokstra/core/request"                  // NOT used in handlers
	core_repository "github.com/primadi/lokstra-auth/infrastructure/repository"  // USED in @Inject
)

// @Handler name="credential-service", prefix="/api"
type CredentialService struct {
	// @Inject "credential-repository"
	Repo core_repository.CredentialRepository
}

// @Route "GET /users/{id}"
func (s *CredentialService) GetUser(id string) (string, error) {
	// Method doesn't use domain or request packages
	return "user-" + id, nil
}
```

**Before fix:**
```go
// zz_generated.lokstra.go
import (
	"github.com/primadi/lokstra-auth/credential/domain"        // ❌ UNUSED
	"github.com/primadi/lokstra/core/request"                  // ❌ UNUSED
	core_repository "github.com/primadi/lokstra-auth/infrastructure/repository"  // ✅ USED
)
```

**After fix:**
```go
// zz_generated.lokstra.go
import (
	core_repository "github.com/primadi/lokstra-auth/infrastructure/repository"  // ✅ Only used imports
)
```

## Root Cause
Bug in `collectPackagesFromType()` function:
```go
// BEFORE (BUGGY):
if start := strings.Index(typeStr, "["); start != -1 {  // ❌ Matches array brackets!
    // This treats []*domain.User as generic type
    // Returns early without extracting "domain" package
}
```

The function was checking for `[` at **any position**, which matched array/slice syntax like `[]*domain.User`, treating it as a generic type and returning early without extracting the package name.

## Solution
Fix the order of operations:
1. **First**: Remove array/pointer prefixes (`*`, `[]`)
2. **Then**: Check for generics (which appear AFTER type name, like `Type[T]`)

```go
// AFTER (FIXED):
// Remove pointer and array prefixes FIRST before checking for generics
cleanType := strings.TrimLeft(typeStr, "*[]")  // []*domain.User -> domain.User

// Handle generics: Type[Param1, Param2]
// Generic brackets appear AFTER the type name, not at the start
if start := strings.Index(cleanType, "["); start != -1 {
    // Now correctly handles: Result[*domain.User], Option[domain.User]
    // And skips: domain.User (no brackets after cleaning)
}
```

## Test Coverage
Added 2 new tests:
1. `TestUnusedImportsNotIncluded` - Verifies unused imports are filtered out
2. `TestOnlyMethodTypesIncluded` - Verifies only handler method types are included

## Impact
- ✅ Cleaner generated code
- ✅ No unused import warnings
- ✅ Faster compilation (fewer imports to resolve)
- ✅ Better IDE performance (fewer packages to index)

## Files Changed
- `core/annotation/codegen.go` - Fixed `collectPackagesFromType()` logic
- `core/annotation/test/unused_imports_test.go` - Added test coverage
