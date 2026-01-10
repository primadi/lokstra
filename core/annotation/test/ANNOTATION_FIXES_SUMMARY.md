# Annotation Parser Fixes & Enhancements

## Overview
This document summarizes three critical fixes to Lokstra's annotation parser to improve accuracy, validation, and cleanup behavior.

---

## 1. Indentation Validation (Documentation vs Code) ✅

### Problem
Annotations dalam dokumentasi (TAB-indented code examples) dianggap sebagai annotation yang valid:

```go
// Example usage in documentation:
//
//	@EndpointService name="example-service"
//	type ExampleService struct {}
```

File `zz_generated.lokstra.go` dibuat meskipun annotation di dalam code example.

### Root Cause
Go documentation convention menggunakan TAB setelah `//` untuk code examples:
- `// @EndpointService` → Real annotation ✅
- `//	@EndpointService` (TAB) → Code example in docs ❌

Parser tidak membedakan antara annotation yang valid dengan code examples.

### Solution
**Modified Files:**
1. `core/annotation/arg_parser.go` - `ParseFileAnnotations()`
2. `core/annotation/complex_processor.go` - `fileContainsRouterService()`

**Detection Rules:**
```go
// Valid annotations (ALLOWED):
// @EndpointService name="user-service"
//@EndpointService name="user-service"

// Invalid annotations (IGNORED):
//	@EndpointService name="user-service"    // TAB after //
//  @EndpointService name="user-service"    // Multiple spaces after //
//   @EndpointService name="user-service"   // Multiple spaces after //
```

**Implementation:**
```go
// Check for TAB or multiple spaces after // (indicates code example)
if len(line) > 2 {
    afterSlashes := line[2:]
    
    // TAB after // → code example
    if len(afterSlashes) > 0 && afterSlashes[0] == '\t' {
        continue
    }
    
    // Multiple spaces after // → code example
    if len(afterSlashes) >= 2 && afterSlashes[0] == ' ' && afterSlashes[1] == ' ' {
        continue
    }
}
```

### Test Coverage
**File:** `core/annotation/arg_parser_indent_test.go`

**Test Cases:**
1. ✅ TAB-indented annotations → Ignored
2. ✅ Valid annotations → Parsed
3. ✅ Multiple empty lines between code → Handled
4. ✅ Few empty lines → Handled

**Test Results:**
```
=== RUN   TestParseFileAnnotations_IgnoreIndentedAnnotations
    arg_parser_indent_test.go:73: Found 4 annotations (correct - 2 valid, 2 TAB-indented ignored)
--- PASS: TestParseFileAnnotations_IgnoreIndentedAnnotations (0.01s)
```

### Two-Step Detection Consistency
Both detection steps now use same logic:

**Step 1:** `fileContainsRouterService()` - Quick check
```go
if strings.Contains(line, "@EndpointService") {
    afterSlashes := line[2:]
    if len(afterSlashes) > 0 && afterSlashes[0] == '\t' {
        continue
    }
    if len(afterSlashes) >= 2 && afterSlashes[0] == ' ' && afterSlashes[1] == ' ' {
        continue
    }
    return true
}
```

**Step 2:** `ParseFileAnnotations()` - Full parsing (same validation)

---

## 2. Struct Validation ✅

### Problem
`@EndpointService` bisa ditulis di atas function, interface, atau type alias:

```go
// @EndpointService name="invalid-service"
func GetUser() {}  // ❌ Invalid tapi tidak error!

// @EndpointService name="invalid-service"
type UserRepository interface {}  // ❌ Invalid tapi tidak error!
```

Annotation di-ignore tanpa error message yang jelas.

### Solution
**Modified:** `core/annotation/codegen.go`

**Added Function:** `isStructDeclaration()`
```go
func isStructDeclaration(fset *token.FileSet, file *ast.File, line int) bool {
    for _, decl := range file.Decls {
        genDecl, ok := decl.(*ast.GenDecl)
        if !ok || genDecl.Tok != token.TYPE {
            continue
        }
        
        for _, spec := range genDecl.Specs {
            typeSpec := spec.(*ast.TypeSpec)
            if fset.Position(typeSpec.Pos()).Line == line+1 {
                _, isStruct := typeSpec.Type.(*ast.StructType)
                return isStruct
            }
        }
    }
    return false
}
```

**Validation in:** `processFileForCodeGen()`
```go
if routerService != nil {
    // Validate that @EndpointService is on a struct
    if !isStructDeclaration(fset, file, routerService.Line) {
        return nil, fmt.Errorf(
            "@EndpointService at line %d must be placed above a struct declaration, "+
            "not a function, interface, or type alias",
            routerService.Line+1,
        )
    }
}
```

### Test Coverage
**File:** `core/annotation/codegen_validation_test.go`

**Test Cases:**
1. ✅ Valid struct → Success
2. ✅ Invalid function → Error: "must be placed above a struct declaration"
3. ✅ Invalid interface → Error: "must be placed above a struct declaration"
4. ✅ Invalid type alias → Error: "must be placed above a struct declaration"

**Test Results:**
```
=== RUN   TestRouterServiceValidation_MustBeOnStruct
=== RUN   TestRouterServiceValidation_MustBeOnStruct/valid_-_struct
=== RUN   TestRouterServiceValidation_MustBeOnStruct/invalid_-_function
=== RUN   TestRouterServiceValidation_MustBeOnStruct/invalid_-_interface
=== RUN   TestRouterServiceValidation_MustBeOnStruct/invalid_-_type_alias
--- PASS: TestRouterServiceValidation_MustBeOnStruct (0.06s)
```

---

## 3. Cleanup Logic Fix ✅

### Problem
File `zz_generated.lokstra.go` tidak auto-delete ketika semua annotations dihapus:
- File `zz_cache.lokstra.json` auto-delete ✅
- File `zz_generated.lokstra.go` tetap ada ❌

### Root Cause Analysis
**Issue:** `GenerateCodeForFolder()` memiliki early return ketika semua files di-skip (karena cache):

```go
// Early return if nothing changed
if len(ctx.UpdatedFiles) == 0 && len(ctx.DeletedFiles) == 0 {
    return nil  // ❌ Skip cleanup logic!
}
```

**Scenario yang bermasalah:**
1. User punya `user_service.go` dengan `@EndpointService` → generated file dibuat ✅
2. User hapus annotation dari `user_service.go`
3. Cache mendeteksi file checksum sama → file di-skip
4. `UpdatedFiles` dan `DeletedFiles` kosong → early return
5. Generated file tidak di-check untuk cleanup ❌

### Solution
**Modified:** `core/annotation/codegen.go` - `GenerateCodeForFolder()`

```go
// Before early return, check if existing generated file should be cleaned up
if len(ctx.UpdatedFiles) == 0 && len(ctx.DeletedFiles) == 0 {
    // ✅ Check if empty generated file exists (orphaned file scenario)
    generatedPath := filepath.Join(ctx.FolderPath, internal.GeneratedFileName)
    if _, err := os.Stat(generatedPath); err == nil {
        // File exists, check if it's empty (no services)
        if len(ctx.GeneratedCode.Services) == 0 {
            if err := os.Remove(generatedPath); err == nil {
                fmt.Fprintf(os.Stderr, "[lokstra-annotation] 🗑️  Deleted empty %s in %s\n", 
                    internal.GeneratedFileName, ctx.FolderPath)
            }
        }
    }
    return nil
}
```

**Logic Flow:**
1. ✅ Early return tetap ada untuk performa
2. ✅ Sebelum return, check jika generated file ada
3. ✅ Jika ada dan kosong (`len(Services) == 0`), delete file
4. ✅ Print message untuk visibility

### Test Coverage
**File:** `core/annotation/codegen_cleanup_test.go`

**Test 1:** `TestGenerateCodeForFolder_CleanupEmptyFile`
- Scenario: User menghapus annotation dari existing file
- Steps:
  1. Create file WITH annotation → generate code → file created ✅
  2. Remove annotation from file → parse again → no annotations
  3. Generate code again → file DELETED ✅

**Test 2:** `TestGenerateCodeForFolder_CleanupWhenSkipped`
- Scenario: Orphaned empty generated file dengan semua files di-skip
- Steps:
  1. Create empty `zz_generated.lokstra.go` manually
  2. Context dengan UpdatedFiles=[], DeletedFiles=[], SkippedFiles=[...]
  3. Generate code → empty file DELETED ✅

### Test Results
```
=== RUN   TestGenerateCodeForFolder_CleanupEmptyFile
    codegen_cleanup_test.go:62: ✓ Step 1: Generated file created
    codegen_cleanup_test.go:118: ✓ Step 2: Generated file deleted
--- PASS: TestGenerateCodeForFolder_CleanupEmptyFile (0.06s)

=== RUN   TestGenerateCodeForFolder_CleanupWhenSkipped
    codegen_cleanup_test.go:174: ✓ Empty generated file deleted when all files skipped
--- PASS: TestGenerateCodeForFolder_CleanupWhenSkipped (0.02s)
```

### Scenarios Covered
1. ✅ **New annotations** → Generate file
2. ✅ **Modified annotations** → Update file
3. ✅ **Removed annotations (with UpdatedFiles)** → Delete file
4. ✅ **Removed annotations (with cache skip)** → Delete file
5. ✅ **Orphaned empty file** → Delete file
6. ✅ **Empty folder** → No action

---

## Summary

### Files Modified
1. `core/annotation/arg_parser.go` - Indentation detection
2. `core/annotation/complex_processor.go` - Consistent quick check
3. `core/annotation/codegen.go` - Struct validation + cleanup logic

### Files Added
1. `core/annotation/arg_parser_indent_test.go` - 4 test cases
2. `core/annotation/codegen_validation_test.go` - 4 test cases
3. `core/annotation/codegen_cleanup_test.go` - 2 test cases
4. `core/annotation/examples/annotation_parsing/` - Examples

### Test Summary
```
Total: 10 new test cases
All tests: PASS ✅
Coverage: Indentation, Validation, Cleanup
```

### Impact
- ✅ Documentation examples tidak lagi trigger code generation
- ✅ Invalid annotation placement sekarang error dengan message yang jelas
- ✅ Orphaned generated files auto-cleanup
- ✅ Consistent behavior antara quick check dan full parsing
- ✅ Better developer experience dengan clear error messages
