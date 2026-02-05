# Fix: Annotation Parser Ignores Code Examples in Documentation

## Problem

The annotation parser was incorrectly detecting `@Handler` (and other annotations) in **documentation examples**, causing false positives and generating unwanted code.

### Example of the Problem

```go
package middleware

// Register is a placeholder for middleware registration
//
// Example @Handler annotation:
//
//	@Handler name="tenant-service", prefix="/api/tenants"
//
// The above is just an EXAMPLE, not an actual annotation!
func Register() {
    // ...
}
```

**Before Fix:** The parser would detect the example annotation and generate code for it.  
**After Fix:** The parser correctly ignores it.

## Solution

Updated the annotation parser to follow **Go documentation conventions**:

### Rules for Valid Annotations

1. **Valid annotation:** `// @Handler` (space or no space after `//`)
   ```go
   // @Handler name="user-service"
   type UserService struct {}
   ```

2. **Invalid annotation (TAB-indented):** `//	@Handler` (TAB after `//`)
   ```go
   // Example:
   //
   //	@Handler name="example"  // ← IGNORED (code example)
   ```

3. **Invalid annotation (multi-space indented):** `//   @Handler`
   ```go
   //   @Handler  // ← IGNORED (indented)
   ```

### Go Documentation Convention

In Go, documentation uses **TAB character** after `//` to indicate code examples:

```go
// Example usage:
//
//	service := NewService()  // ← Code example (TAB-indented)
//	result := service.Do()
```

The parser now respects this convention and **skips any annotation that is TAB-indented**.

## Changes Made

### Files Modified

1. **`core/annotation/arg_parser.go`**
   - Updated `ParseFileAnnotations()` to detect TAB-indented annotations
   - Rejects annotations with TAB immediately after `//`
   - Rejects annotations with multiple spaces after `//`
   - Allows single space after `//` (normal comment formatting)

2. **`core/annotation/complex_processor.go`**
   - Updated `fileContainsRouterService()` with same logic
   - Ensures consistency between quick check and full parsing

### Tests Added

1. **`TestParseFileAnnotations_IgnoreIndentedAnnotations`**
   - Verifies TAB-indented annotations are ignored
   - Tests real-world documentation example

2. **`TestParseFileAnnotations_ValidAnnotations`**
   - Ensures valid annotations are still detected
   - Tests `@Handler`, `@Inject`, `@Route`

3. **`TestParseFileAnnotations_MultipleEmptyLinesAfterAnnotation`**
   - Prevents matching annotations with too many gap lines
   - Helps avoid false positives in long documentation blocks

4. **`TestParseFileAnnotations_AnnotationWithFewEmptyLines`**
   - Allows normal documentation structure (1-3 empty comment lines)

5. **`TestDebugIndentDetection`** (debug helper)
   - Helps visualize whitespace detection logic

## Impact

### Before

File with documentation example:
```go
// Example:
//	@Handler name="tenant-service", prefix="/api/tenants"
func Register() {}
```

Generated `zz_generated.lokstra.go` with **unwanted code** for `Register` function.

### After

Same file: **No code generated** (correctly ignored).

Only actual annotations are processed:
```go
// @Handler name="tenant-service", prefix="/api/tenants"
type TenantService struct {}
```

## Backward Compatibility

✅ All existing valid annotations still work  
✅ Only TAB-indented and heavily-indented annotations are now filtered out  
✅ All existing tests pass

## Testing

Run tests:
```bash
cd core/annotation
go test -v
```

All tests should pass:
- `TestParseFileAnnotations_IgnoreIndentedAnnotations` ✅
- `TestParseFileAnnotations_ValidAnnotations` ✅
- `TestParseFileAnnotations_MultipleEmptyLinesAfterAnnotation` ✅
- `TestParseFileAnnotations_AnnotationWithFewEmptyLines` ✅
- `TestDebugIndentDetection` ✅

## Recommendations

When writing documentation with annotation examples:

1. **Use TAB indentation** for code examples (Go convention):
   ```go
   // Example:
   //
   //	@Handler name="example"  // ← Will be ignored
   ```

2. **Or use enough indentation** (2+ spaces):
   ```go
   // Example:
   //   @Handler name="example"  // ← Will be ignored
   ```

3. **For actual annotations**, use no indentation or single space:
   ```go
   // @Handler name="real-service"  // ← Will be processed
   type RealService struct {}
   ```

## Summary

This fix ensures that Lokstra's annotation parser follows Go documentation conventions and doesn't generate code from documentation examples. The parser is now smarter about distinguishing between real annotations and code examples.
