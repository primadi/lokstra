package annotation

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/primadi/lokstra/core/annotation/internal"
)

// GenerateCodeForFolder generates zz_generated.lokstra.go based on RouterServiceContext
func GenerateCodeForFolder(ctx *RouterServiceContext) error {
	genPath := filepath.Join(ctx.FolderPath, internal.GeneratedFileName)

	// If all files are skipped and no files were deleted, skip generation
	// BUT still check if we need to cleanup empty generated file
	if len(ctx.UpdatedFiles) == 0 && len(ctx.DeletedFiles) == 0 {
		// Check if existing generated file has no services
		// This handles case where annotations were removed but file was skipped due to cache
		existingGenCode := readExistingGenCode(genPath)
		if len(existingGenCode) == 0 {
			// No services in existing file, remove it
			os.Remove(genPath)
		}
		return nil
	}

	// Read existing zz_generated.lokstra.go to preserve code for skipped files
	existingGenCode := readExistingGenCode(genPath)
	existingImports := extractImportsFromExistingGenCode(genPath)

	// Process all updated files
	for _, file := range ctx.UpdatedFiles {
		if err := processFileForCodeGen(file, ctx); err != nil {
			return fmt.Errorf("failed to process %s: %w", file.Filename, err)
		}
	}

	// For skipped files, copy existing code from zz_generated.lokstra.go
	for _, file := range ctx.SkippedFiles {
		if code, exists := existingGenCode[file.Filename]; exists {
			// Store the existing code section
			ctx.GeneratedCode.PreservedSections[file.Filename] = code
		}
	}

	// Generate zz_generated.lokstra.go (pass existing imports for preservation)
	if len(ctx.GeneratedCode.Services) > 0 || len(ctx.GeneratedCode.PreservedSections) > 0 {
		if err := writeGenFile(genPath, ctx, existingImports); err != nil {
			return fmt.Errorf("failed to write zz_generated.lokstra.go: %w", err)
		}
	} else {
		// No services, remove zz_generated.lokstra.go if exists
		os.Remove(genPath)
	}

	return nil
}

// readExistingGenCode reads existing zz_generated.lokstra.go and extracts code sections per file
func readExistingGenCode(genPath string) map[string]string {
	sections := make(map[string]string)

	content, err := os.ReadFile(genPath)
	if err != nil {
		return sections
	}

	text := string(content)

	// Split by file separator
	// Pattern: // ============================================================
	//          // FILE: user_service.go
	//          // ============================================================

	lines := strings.Split(text, "\n")
	var currentFile string
	var currentSection strings.Builder
	inSection := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check for file separator start
		if strings.Contains(line, "// ============================================================") {
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "// FILE: ") {
				// Save previous section if exists
				if inSection && currentFile != "" {
					sections[currentFile] = currentSection.String()
				}

				// Start new section
				currentFile = strings.TrimPrefix(lines[i+1], "// FILE: ")
				currentFile = strings.TrimSpace(currentFile)
				currentSection.Reset()

				// Skip the separator lines
				currentSection.WriteString(line + "\n")
				currentSection.WriteString(lines[i+1] + "\n")
				if i+2 < len(lines) {
					currentSection.WriteString(lines[i+2] + "\n")
					i += 2
				}

				inSection = true
				continue
			}
		}

		// Add line to current section
		if inSection {
			currentSection.WriteString(line + "\n")
		}
	}

	// Save last section
	if inSection && currentFile != "" {
		sections[currentFile] = currentSection.String()
	}

	return sections
}

// extractImportsFromExistingGenCode extracts imports from existing zz_generated.lokstra.go
// Returns map[alias]importPath for all imports that were previously used
func extractImportsFromExistingGenCode(genPath string) map[string]string {
	imports := make(map[string]string)

	content, err := os.ReadFile(genPath)
	if err != nil {
		return imports
	}

	// Parse the existing generated file
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, genPath, content, parser.ImportsOnly)
	if err != nil {
		return imports
	}

	// Extract imports
	for _, imp := range astFile.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		var alias string
		if imp.Name != nil {
			// Named import: import foo "path/to/foo"
			alias = imp.Name.Name
		} else {
			// Default import: extract last part of path
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}

		imports[alias] = importPath
	}

	return imports
}

// extractImports extracts import statements from source file
func extractImports(astFile *ast.File, service *ServiceGeneration) error {
	for _, imp := range astFile.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		var alias string
		if imp.Name != nil {
			// Named import: import foo "path/to/foo"
			alias = imp.Name.Name
		} else {
			// Default import: extract last part of path
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}

		service.Imports[alias] = importPath
	}

	return nil
}

// processFileForCodeGen processes a single file for code generation
func processFileForCodeGen(file *FileToProcess, ctx *RouterServiceContext) error {
	// Parse the file to get AST
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file.FullPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Find @RouterService and @Service annotations
	for _, ann := range file.Annotations {
		if ann.Name != "RouterService" && ann.Name != "Service" {
			continue
		}

		isService := ann.Name == "Service"

		// Read common args
		var serviceName string
		var prefix string
		var middlewares []string

		if isService {
			// @Service only needs name
			args, err := ann.ReadArgs("name")
			if err != nil {
				return fmt.Errorf("@Service on line %d: %w", ann.Line, err)
			}
			serviceName, _ = args["name"].(string)
		} else {
			// @RouterService needs name, prefix, middlewares
			args, err := ann.ReadArgs("name", "prefix", "middlewares")
			if err != nil {
				return fmt.Errorf("@RouterService on line %d: %w", ann.Line, err)
			}
			serviceName, _ = args["name"].(string)
			prefix, _ = args["prefix"].(string)
			middlewares = extractStringArray(args["middlewares"])
		}

		if serviceName == "" {
			return fmt.Errorf("@%s on line %d: 'name' is required", ann.Name, ann.Line)
		}

		// VALIDATE: must be placed above a struct declaration
		if !isStructDeclaration(astFile, ann.TargetName) {
			return fmt.Errorf("@%s on line %d: must be placed directly above a struct declaration, found '%s' instead (file: %s)",
				ann.Name, ann.Line, ann.TargetName, file.Filename)
		}

		// Create service generation entry
		service := &ServiceGeneration{
			ServiceName:        serviceName,
			Prefix:             prefix,
			Middlewares:        middlewares,
			Routes:             make(map[string]string),
			RouteMiddlewares:   make(map[string][]string),
			Methods:            make(map[string]*MethodSignature),
			Dependencies:       make(map[string]*DependencyInfo),
			ConfigDependencies: make(map[string]*ConfigInfo),
			Imports:            make(map[string]string),
			StructName:         ann.TargetName,
			SourceFile:         file.Filename,
			IsService:          isService,
		}

		// Extract imports from source file
		if err := extractImports(astFile, service); err != nil {
			return err
		}

		// Find interface name and methods
		if err := extractInterfaceInfo(astFile, ann.TargetName, service); err != nil {
			return err
		}

		// Extract method signatures from struct
		if err := extractMethodSignatures(astFile, ann.TargetName, service); err != nil {
			return err
		}

		// Find @Route annotations on methods
		if err := extractRoutes(file, service); err != nil {
			return err
		}

		// Find @Inject annotations for dependencies
		if err := extractDependencies(file, service); err != nil {
			return err
		}

		// Find @InjectCfg annotations for config dependencies
		if err := extractConfigDependencies(file, service); err != nil {
			return err
		}

		// Check if struct has Init() error method
		checkInitMethod(file, service)

		ctx.GeneratedCode.Services[serviceName] = service
	}

	return nil
}

// isStructDeclaration checks if a name refers to a struct type declaration in the AST
func isStructDeclaration(astFile *ast.File, name string) bool {
	if name == "" {
		return false
	}

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// Check if this is the type we're looking for
			if typeSpec.Name.Name == name {
				// Check if it's a struct type
				_, isStruct := typeSpec.Type.(*ast.StructType)
				return isStruct
			}
		}
	}

	return false
}

// extractInterfaceInfo finds the interface that the struct implements
func extractInterfaceInfo(astFile *ast.File, structName string, service *ServiceGeneration) error {
	// Look for: var _ InterfaceName = (*StructName)(nil)
	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Check if it's interface assertion
			if len(valueSpec.Values) > 0 {
				// Look for (*StructName)(nil)
				if call, ok := valueSpec.Values[0].(*ast.CallExpr); ok {
					if star, ok := call.Fun.(*ast.StarExpr); ok {
						if ident, ok := star.X.(*ast.Ident); ok && ident.Name == structName {
							// Found it! Extract interface name
							if len(valueSpec.Names) > 0 && valueSpec.Names[0].Name == "_" {
								if ident, ok := valueSpec.Type.(*ast.Ident); ok {
									service.InterfaceName = ident.Name
									service.RemoteTypeName = ident.Name + "Remote"
									return nil
								}
							}
						}
					}
				}
			}
		}
	}

	// Default if not found
	service.InterfaceName = structName + "Interface"
	service.RemoteTypeName = structName + "Remote"
	return nil
}

// extractMethodSignatures extracts method signatures from struct methods
func extractMethodSignatures(astFile *ast.File, structName string, service *ServiceGeneration) error {
	for _, decl := range astFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			continue
		}

		// Check if method belongs to our struct
		if len(funcDecl.Recv.List) == 0 {
			continue
		}

		recvType := funcDecl.Recv.List[0].Type
		var recvName string

		// Handle *StructName or StructName
		if starExpr, ok := recvType.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				recvName = ident.Name
			}
		} else if ident, ok := recvType.(*ast.Ident); ok {
			recvName = ident.Name
		}

		// Skip if method doesn't belong to our target struct
		if recvName != structName {
			continue
		}

		// Extract method signature
		methodName := funcDecl.Name.Name
		sig := &MethodSignature{
			Name: methodName,
		}

		// Extract parameter type (assume single param)
		// Skip if no parameters or only receiver
		if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
			firstParam := funcDecl.Type.Params.List[0]

			// Only set ParamType if there are actual named parameters with types
			if len(firstParam.Names) > 0 && firstParam.Type != nil {
				sig.ParamType = exprToString(firstParam.Type)
			}
		}
		if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
			numResults := len(funcDecl.Type.Results.List)

			switch numResults {
			case 1:
				// Only error
				sig.HasData = false
			case 2:
				// (T, error)
				sig.ReturnType = exprToString(funcDecl.Type.Results.List[0].Type)
				sig.HasData = true
			}
		}

		service.Methods[methodName] = sig
	}

	return nil
}

// exprToString converts ast.Expr to string representation
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "any"
	case *ast.IndexExpr:
		// Generic type: Type[T]
		return exprToString(t.X) + "[" + exprToString(t.Index) + "]"
	case *ast.IndexListExpr:
		// Generic type with multiple params: Type[T1, T2]
		var typeArgs []string
		for _, arg := range t.Indices {
			typeArgs = append(typeArgs, exprToString(arg))
		}
		return exprToString(t.X) + "[" + strings.Join(typeArgs, ", ") + "]"
	default:
		return "any"
	}
}

// extractRoutes finds all @Route annotations on methods
func extractRoutes(file *FileToProcess, service *ServiceGeneration) error {
	for _, ann := range file.Annotations {
		if ann.Name != "Route" && ann.Name != "route" {
			continue
		}

		// Supported formats (DO NOT MIX positional and named!):
		// 1. @Route "GET /users/{id}"                            - positional only
		// 2. @Route route="GET /users/{id}", middlewares=[...]   - named only

		var routeStr string
		var middlewares []string

		// Try to read args
		args, err := ann.ReadArgs("route", "middlewares")
		if err != nil {
			return fmt.Errorf("@Route on line %d: %w", ann.Line, err)
		}

		// Extract route
		routeStr, _ = args["route"].(string)

		// Extract middlewares
		middlewares = extractStringArray(args["middlewares"])

		if routeStr == "" {
			return fmt.Errorf(`@Route on line %d: route string is required. Valid formats:
  - Positional only: @Route "GET /path"
  - Named only: @Route route="GET /path", middlewares=["auth"]
Note: Cannot mix positional and named arguments`, ann.Line)
		}

		// Remove query parameters if any
		// Query params are not part of route mapping, only path is needed
		idxQuestion := strings.Index(routeStr, "?")
		if idxQuestion != -1 {
			routeStr = routeStr[:idxQuestion]
		}

		if ann.TargetName != "" {
			service.Routes[ann.TargetName] = routeStr

			// Store route middlewares if present
			if len(middlewares) > 0 {
				service.RouteMiddlewares[ann.TargetName] = middlewares
			}
		}
	}

	return nil
}

// extractDependencies finds all @Inject annotations and field info
func extractDependencies(file *FileToProcess, service *ServiceGeneration) error {

	// Parse file to get field types
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file.FullPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Find struct fields for THIS struct only
	fieldTypes := make(map[string]string)     // fieldName -> fieldType
	structFieldNames := make(map[string]bool) // fieldName -> true (for THIS struct only)

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != service.StructName {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Extract field names and types ONLY for THIS struct
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue
				}
				fieldName := field.Names[0].Name
				fieldType := exprToString(field.Type)
				fieldTypes[fieldName] = fieldType
				structFieldNames[fieldName] = true
			}
			// Break after finding the target struct to avoid processing other structs
			break
		}
	}

	// Now process @Inject annotations - ONLY for fields belonging to THIS struct
	for _, ann := range file.Annotations {
		if ann.Name != "Inject" && ann.Name != "inject" {
			continue
		}

		// Skip if annotation target is not a field of THIS struct
		if !structFieldNames[ann.TargetName] {
			continue
		}

		// Supported formats:
		// @Inject "user-repository"
		// @Inject service="user-repository"
		args, err := ann.ReadArgs("service")
		if err != nil {
			return fmt.Errorf("@Inject on line %d: %w", ann.Line, err)
		}

		var serviceName string
		if svc, ok := args["service"].(string); ok {
			serviceName = svc
		}

		if serviceName != "" && ann.TargetName != "" {
			// ann.TargetName is field name
			fieldType := fieldTypes[ann.TargetName]

			service.Dependencies[serviceName] = &DependencyInfo{
				ServiceName: serviceName,
				FieldName:   ann.TargetName,
				FieldType:   fieldType,
			}
		}
	}

	return nil
}

// checkInitMethod checks if struct has Init() error method
func checkInitMethod(file *FileToProcess, service *ServiceGeneration) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file.FullPath, nil, parser.ParseComments)
	if err != nil {
		return
	}

	// Look for method: func (receiver *StructName) Init() error
	for _, decl := range astFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		// Check if method name is Init
		if funcDecl.Name.Name != "Init" {
			continue
		}

		// Check receiver type
		recvType := funcDecl.Recv.List[0].Type
		var receiverName string
		switch t := recvType.(type) {
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				receiverName = ident.Name
			}
		case *ast.Ident:
			receiverName = t.Name
		}

		if receiverName != service.StructName {
			continue
		}

		// Check signature: no params, returns error
		if funcDecl.Type.Params.NumFields() != 0 {
			continue
		}

		if funcDecl.Type.Results == nil || funcDecl.Type.Results.NumFields() != 1 {
			continue
		}

		// Check return type is error
		returnType := funcDecl.Type.Results.List[0].Type
		if ident, ok := returnType.(*ast.Ident); ok && ident.Name == "error" {
			service.HasInitMethod = true
			return
		}
	}
}

// extractConfigDependencies finds all @InjectCfg annotations and field info
func extractConfigDependencies(file *FileToProcess, service *ServiceGeneration) error {
	// Parse file to get field types
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file.FullPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Find struct fields for THIS struct only
	fieldTypes := make(map[string]string)     // fieldName -> fieldType
	structFieldNames := make(map[string]bool) // fieldName -> true (for THIS struct only)

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != service.StructName {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Extract field names and types ONLY for THIS struct
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue
				}
				fieldName := field.Names[0].Name
				fieldType := exprToString(field.Type)
				fieldTypes[fieldName] = fieldType
				structFieldNames[fieldName] = true
			}
			// Break after finding the target struct to avoid processing other structs
			break
		}
	}

	// Now process @InjectCfg annotations - ONLY for fields belonging to THIS struct
	for _, ann := range file.Annotations {
		if ann.Name != "InjectCfg" && ann.Name != "injectcfg" {
			continue
		}

		// Skip if annotation target is not a field of THIS struct
		if !structFieldNames[ann.TargetName] {
			continue
		}

		// Supported formats:
		// @InjectCfg "app.jwt-secret"
		// @InjectCfg key="app.jwt-secret"
		// @InjectCfg key="app.jwt-secret", default="secret"
		// @InjectCfg "app.timeout", "30"  (positional: key, default)
		args, err := ann.ReadArgs("key", "default")
		if err != nil {
			return fmt.Errorf("@InjectCfg on line %d: %w", ann.Line, err)
		}

		var configKey string
		if key, ok := args["key"].(string); ok {
			configKey = key
		}

		// Parse default value (optional) - can be string, int, bool, or float
		defaultValue := ""
		if def, ok := args["default"]; ok && def != nil {
			// Convert any type to string
			switch v := def.(type) {
			case string:
				defaultValue = v
			case int:
				defaultValue = fmt.Sprintf("%d", v)
			case bool:
				defaultValue = fmt.Sprintf("%t", v)
			case float64:
				defaultValue = fmt.Sprintf("%g", v)
			default:
				defaultValue = fmt.Sprintf("%v", v)
			}
		}

		if configKey != "" && ann.TargetName != "" {
			// ann.TargetName is field name
			fieldType := fieldTypes[ann.TargetName]

			service.ConfigDependencies[configKey] = &ConfigInfo{
				ConfigKey:    configKey,
				FieldName:    ann.TargetName,
				FieldType:    fieldType,
				DefaultValue: defaultValue,
			}
		}
	}

	return nil
}

// writeGenFile writes the zz_generated.lokstra.go file
func writeGenFile(path string, ctx *RouterServiceContext, existingImports map[string]string) error {
	// Get package name from existing files
	pkgName, err := getPackageName(ctx.FolderPath)
	if err != nil {
		return err
	}

	// Collect ALL struct names for init() registration (from both updated and skipped files)
	allStructNames := make([]string, 0)
	for _, service := range ctx.GeneratedCode.Services {
		allStructNames = append(allStructNames, service.StructName)
	}
	// Also extract struct names from preserved sections
	for _, code := range ctx.GeneratedCode.PreservedSections {
		// Extract Register<StructName>() calls to get struct names
		if structName := extractStructNameFromPreservedCode(code); structName != "" {
			// Only add if not already in list
			found := false
			for _, existing := range allStructNames {
				if existing == structName {
					found = true
					break
				}
			}
			if !found {
				allStructNames = append(allStructNames, structName)
			}
		}
	}

	// Sort struct names for deterministic order in init()
	sort.Strings(allStructNames)

	// Collect used packages from method signatures, dependencies, and struct name
	usedPackages := make(map[string]bool)
	for _, service := range ctx.GeneratedCode.Services {
		// From method signatures
		for _, method := range service.Methods {
			collectPackagesFromType(method.ParamType, usedPackages)
			collectPackagesFromType(method.ReturnType, usedPackages)
		}
		// From dependencies
		for _, dep := range service.Dependencies {
			collectPackagesFromType(dep.FieldType, usedPackages)
		}
		// From config dependencies - check if time.Duration is used
		for _, cfg := range service.ConfigDependencies {
			if cfg.FieldType == "time.Duration" {
				usedPackages["time"] = true
			}
		}
	}

	// Also collect packages from preserved sections (unchanged files)
	for _, code := range ctx.GeneratedCode.PreservedSections {
		extractPackagesFromCode(code, usedPackages)
	}

	// Filter imports to only used packages
	allImports := make(map[string]string) // path -> alias

	// Hardcoded imports that are always included in template
	hardcodedImports := map[string]bool{
		"github.com/primadi/lokstra/core/deploy":      true,
		"github.com/primadi/lokstra/core/proxy":       true,
		"github.com/primadi/lokstra/lokstra_registry": true,
	}

	// First, add imports from updated services
	for _, service := range ctx.GeneratedCode.Services {
		for alias, importPath := range service.Imports {
			// Skip hardcoded imports
			if hardcodedImports[importPath] {
				continue
			}
			// Only include if package is actually used in generated code
			if usedPackages[alias] {
				// Use path as key to deduplicate, prefer shorter alias
				if existing, exists := allImports[importPath]; !exists || len(alias) < len(existing) {
					allImports[importPath] = alias
				}
			}
		}
	}

	// Second, add imports from existing generated file if package is still used
	for alias, importPath := range existingImports {
		// Skip hardcoded imports
		if hardcodedImports[importPath] {
			continue
		}
		if usedPackages[alias] {
			// Use path as key to deduplicate, prefer shorter alias
			if existing, exists := allImports[importPath]; !exists || len(alias) < len(existing) {
				allImports[importPath] = alias
			}
		}
	}

	// Third, add "time" if time.Duration is used
	if usedPackages["time"] {
		allImports["time"] = "time"
	}

	// Generate code
	var buf bytes.Buffer
	if err := genTemplate.Execute(&buf, map[string]any{
		"Package":           pkgName,
		"Services":          ctx.GeneratedCode.Services,
		"PreservedSections": ctx.GeneratedCode.PreservedSections,
		"AllImports":        allImports,
		"AllStructNames":    allStructNames,
	}); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// extractStructNameFromPreservedCode extracts struct name from Register<StructName>() function
func extractStructNameFromPreservedCode(code string) string {
	// Look for pattern: func Register<StructName>()
	if idx := strings.Index(code, "func Register"); idx != -1 {
		start := idx + len("func Register")
		end := strings.Index(code[start:], "()")
		if end != -1 {
			return code[start : start+end]
		}
	}
	return ""
}

// extractPackagesFromCode scans preserved code for package usages
// This finds patterns like: domain.User, *domain.User, []domain.User, pkg.Type, etc.
func extractPackagesFromCode(code string, packages map[string]bool) {
	// Split code into lines and scan for type references
	// Look for patterns: pkg.Type where pkg starts with lowercase and Type starts with uppercase
	lines := strings.Split(code, "\n")

	// Regex to match package.Type patterns
	// Matches qualified identifiers like: authdomain.LoginRequest, *userdomain.UserDTO, etc.
	re := regexp.MustCompile(`([a-z][a-zA-Z0-9_]*)\.([A-Z][a-zA-Z0-9_]*)`)

	for _, line := range lines {
		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				pkg := match[1]
				packages[pkg] = true
			}
		}
	}
}

// collectPackagesFromType extracts package prefixes from type string
// e.g., "*domain.User" -> "domain", "[]userDomain.User" -> "userDomain"
func collectPackagesFromType(typeStr string, packages map[string]bool) {
	if typeStr == "" {
		return
	}

	// Handle generics first: Type[Param1, Param2]
	if start := strings.Index(typeStr, "["); start != -1 {
		if end := strings.LastIndex(typeStr, "]"); end > start {
			// Process the base type (before '[')
			baseType := typeStr[:start]
			collectPackagesFromType(baseType, packages)

			// Process inner type parameters
			innerTypes := typeStr[start+1 : end]
			// Split by comma for multiple type params
			for _, inner := range strings.Split(innerTypes, ",") {
				collectPackagesFromType(strings.TrimSpace(inner), packages)
			}
			return
		}
	}

	// Remove pointer and array prefixes
	typeStr = strings.TrimLeft(typeStr, "*[]")

	// Extract package prefix (everything before last dot)
	if idx := strings.LastIndex(typeStr, "."); idx != -1 {
		pkg := typeStr[:idx]
		// Handle nested packages (e.g., "github.com/user/repo.Type" -> "repo")
		// Only take the last segment as the alias
		if lastSlash := strings.LastIndex(pkg, "/"); lastSlash != -1 {
			pkg = pkg[lastSlash+1:]
		}
		packages[pkg] = true
	}
}

// getPackageName gets package name from a folder
func getPackageName(folderPath string) (string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(file.Name(), "_test.go") || file.Name() == internal.GeneratedFileName {
			continue
		}

		fullPath := filepath.Join(folderPath, file.Name())
		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, fullPath, nil, parser.PackageClauseOnly)
		if err != nil {
			continue
		}

		return astFile.Name.Name, nil
	}

	return "application", nil
}

// extractStringArray extracts string array from any
func extractStringArray(val any) []string {
	if val == nil {
		return []string{}
	}

	if arr, ok := val.([]string); ok {
		return arr
	}

	if str, ok := val.(string); ok {
		// Split by comma
		parts := strings.Split(str, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if p := strings.TrimSpace(part); p != "" {
				result = append(result, p)
			}
		}
		return result
	}

	return []string{}
}

// convertDurationToGo converts duration string like "24h" to Go duration expression "24*time.Hour"
func convertDurationToGo(durationStr string) string {
	if durationStr == "" {
		return "0"
	}

	// Parse duration value and unit
	var value string
	var unit string

	// Find where the number ends and unit begins
	for i, ch := range durationStr {
		if ch >= '0' && ch <= '9' || ch == '.' {
			continue
		}
		value = durationStr[:i]
		unit = durationStr[i:]
		break
	}

	if value == "" || unit == "" {
		return "0"
	}

	// Map unit to Go time constant
	var goUnit string
	switch unit {
	case "ns":
		goUnit = "time.Nanosecond"
	case "us", "µs":
		goUnit = "time.Microsecond"
	case "ms":
		goUnit = "time.Millisecond"
	case "s":
		goUnit = "time.Second"
	case "m":
		goUnit = "time.Minute"
	case "h":
		goUnit = "time.Hour"
	default:
		return "0"
	}

	return fmt.Sprintf("%s*%s", value, goUnit)
}

// genTemplate is the template for zz_generated.lokstra.go
var genTemplate = template.Must(template.New("gen").Funcs(template.FuncMap{
	"hasReturnValue":    hasReturnValue,
	"extractReturnType": extractReturnType,
	"quote":             func(s string) string { return fmt.Sprintf("%q", s) },
	"join":              strings.Join,
	"trimPrefix":        strings.TrimPrefix,
	"trimSuffix":        strings.TrimSuffix,
	"notEmpty":          func(s string) bool { return strings.TrimSpace(s) != "" },

	"getDefaultValue": func(fieldType, defaultValue string) string {
		if defaultValue != "" {
			// For string type, add quotes if not already quoted
			if fieldType == "string" {
				if !strings.HasPrefix(defaultValue, `"`) {
					return fmt.Sprintf(`"%s"`, defaultValue)
				}
				return defaultValue
			}

			// For duration type, parse and convert to proper Go syntax
			if fieldType == "time.Duration" {
				// defaultValue like "24h", "5m", "30s"
				// Need to convert to Go duration expression
				return convertDurationToGo(defaultValue)
			} // For int/bool/float, return as-is
			return defaultValue
		}
		// Generate type-specific zero values
		switch fieldType {
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			return "0"
		case "bool":
			return "false"
		case "float32", "float64":
			return "0.0"
		case "time.Duration":
			return "0"
		default:
			return `""`
		}
	},
	"sortedKeys": func(m any) []string {
		keys := make([]string, 0)
		switch v := m.(type) {
		case map[string]string:
			for k := range v {
				keys = append(keys, k)
			}
		case map[string][]string:
			for k := range v {
				keys = append(keys, k)
			}
		case map[string]*DependencyInfo:
			for k := range v {
				keys = append(keys, k)
			}
		case map[string]*ConfigInfo:
			for k := range v {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		return keys
	},
}).Parse(`// AUTO-GENERATED CODE - DO NOT EDIT
// Generated by lokstra-annotation from annotations in this folder
// Annotations: @RouterService, @Service, @Inject, @InjectCfg, @Route

package {{.Package}}

import (
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/lokstra_registry"
{{- range $path, $alias := .AllImports }}
	{{$alias}} "{{$path}}"
{{- end }}
)

// Auto-register on package import
func init() {
{{- range $structName := .AllStructNames }}
	Register{{$structName}}()
{{- end }}
}
{{range $name, $service := .Services}}
// ============================================================
// FILE: {{$service.SourceFile}}
// ============================================================
{{if $service.IsService}}
// Register{{$service.StructName}} registers the {{$service.ServiceName}} with the registry
// Auto-generated from annotations:
//   - @Service name={{quote $service.ServiceName}}
{{- if $service.Dependencies}}
//   - @Inject annotations
{{- end}}
{{- if $service.ConfigDependencies}}
//   - @InjectCfg annotations
{{- end}}
func Register{{$service.StructName}}() {
	lokstra_registry.RegisterLazyService({{quote $service.ServiceName}}, func(deps map[string]any, cfg map[string]any) any {
		svc := &{{$service.StructName}}{
{{- range $key := sortedKeys $service.Dependencies }}
{{- $dep := index $service.Dependencies $key }}
			{{$dep.FieldName}}: deps[{{quote $dep.ServiceName}}].({{$dep.FieldType}}),
{{- end }}
{{- range $key := sortedKeys $service.ConfigDependencies }}
{{- $cfg := index $service.ConfigDependencies $key }}
			{{$cfg.FieldName}}: cfg[{{quote $cfg.ConfigKey}}].({{$cfg.FieldType}}),
{{- end }}
		}
{{- if $service.HasInitMethod }}
		
		// Call Init() for post-initialization
		if err := svc.Init(); err != nil {
			panic("failed to initialize {{$service.ServiceName}}: " + err.Error())
		}
{{- end }}
		
		return svc
	}, map[string]any{
{{- if or $service.Dependencies $service.ConfigDependencies }}
{{- if $service.Dependencies }}
		"depends-on": []string{ {{range $key := sortedKeys $service.Dependencies}}{{$dep := index $service.Dependencies $key}}{{quote $dep.ServiceName}}, {{end}}},
{{- end }}
{{- range $key := sortedKeys $service.ConfigDependencies }}
{{- $cfg := index $service.ConfigDependencies $key }}
		{{quote $cfg.ConfigKey}}: lokstra_registry.GetConfig({{quote $cfg.ConfigKey}}, {{getDefaultValue $cfg.FieldType $cfg.DefaultValue}}),
{{- end }}
{{- end }}
	})
}
{{else}}
// {{$service.RemoteTypeName}} implements {{$service.InterfaceName}} with HTTP proxy
// Auto-generated from {{$service.StructName}} interface methods
type {{$service.RemoteTypeName}} struct {
	proxyService *proxy.Service
}

// New{{$service.RemoteTypeName}} creates a new remote {{$service.ServiceName}} proxy
func New{{$service.RemoteTypeName}}(proxyService *proxy.Service) *{{$service.RemoteTypeName}} {
	return &{{$service.RemoteTypeName}}{
		proxyService: proxyService,
	}
}

{{range $method := sortedKeys $service.Routes}}
{{- $route := index $service.Routes $method}}
{{- $sig := index $service.Methods $method}}
{{- if $sig}}// {{$method}} via HTTP
// Generated from: @Route {{quote $route}}
{{if $sig.HasData}}{{if notEmpty $sig.ParamType}}func (s *{{$service.RemoteTypeName}}) {{$method}}(p {{$sig.ParamType}}) ({{$sig.ReturnType}}, error) {
	return proxy.CallWithData[{{$sig.ReturnType}}](s.proxyService, {{quote $method}}, p)
}
{{else}}func (s *{{$service.RemoteTypeName}}) {{$method}}() ({{$sig.ReturnType}}, error) {
	return proxy.CallWithData[{{$sig.ReturnType}}](s.proxyService, {{quote $method}}, nil)
}
{{end}}{{else}}{{if notEmpty $sig.ParamType}}func (s *{{$service.RemoteTypeName}}) {{$method}}(p {{$sig.ParamType}}) error {
	return proxy.Call(s.proxyService, {{quote $method}}, p)
}
{{else}}func (s *{{$service.RemoteTypeName}}) {{$method}}() error {
	return proxy.Call(s.proxyService, {{quote $method}}, nil)
}{{end}}{{end}}
{{- end}}
{{end}}
func {{$service.StructName}}Factory(deps map[string]any, config map[string]any) any {
	svc := &{{$service.StructName}}{
{{- range $key := sortedKeys $service.Dependencies }}
{{- $dep := index $service.Dependencies $key }}
		{{$dep.FieldName}}: deps[{{quote $dep.ServiceName}}].({{$dep.FieldType}}),
{{- end }}
{{- range $key := sortedKeys $service.ConfigDependencies }}
{{- $cfg := index $service.ConfigDependencies $key }}
		{{$cfg.FieldName}}: config[{{quote $cfg.ConfigKey}}].({{$cfg.FieldType}}),
{{- end }}
	}
{{- if $service.HasInitMethod }}
	
	// Call Init() for post-initialization
	if err := svc.Init(); err != nil {
		panic("failed to initialize {{$service.ServiceName}}: " + err.Error())
	}
{{- end }}
	
	return svc
}

// {{$service.RemoteTypeName}}Factory creates a remote HTTP client for {{$service.InterfaceName}}
// Auto-generated from @RouterService annotation
func {{$service.RemoteTypeName}}Factory(deps, config map[string]any) any {
	proxyService, ok := config["remote"].(*proxy.Service)
	if !ok {
		panic("remote factory requires 'remote' (proxy.Service) in config")
	}
	return New{{$service.RemoteTypeName}}(proxyService)
}

// Register{{$service.StructName}} registers the {{$service.ServiceName}} with the registry
// Auto-generated from annotations:
//   - @RouterService name={{quote $service.ServiceName}}, prefix={{quote $service.Prefix}}
{{- if $service.Dependencies}}
//   - @Inject annotations
{{- end}}
{{- if $service.ConfigDependencies}}
//   - @InjectCfg annotations
{{- end}}
//   - @Route annotations on methods
func Register{{$service.StructName}}() {
	// Register service type with router configuration
	lokstra_registry.RegisterRouterServiceType("{{$service.ServiceName}}-factory",
		{{$service.StructName}}Factory,
		{{$service.RemoteTypeName}}Factory,
		&deploy.ServiceTypeConfig{
			PathPrefix:  {{quote $service.Prefix}},
			Middlewares: []string{ {{range $i, $mw := $service.Middlewares}}{{if $i}}, {{end}}{{quote $mw}}{{end}} },
			RouteOverrides: map[string]deploy.RouteConfig{
{{- range $method := sortedKeys $service.Routes }}
{{- $route := index $service.Routes $method }}
{{- $middlewares := index $service.RouteMiddlewares $method }}
				{{quote $method}}: {
					Path: {{quote $route}},
{{- if $middlewares }}
					Middlewares: []string{ {{range $i, $mw := $middlewares}}{{if $i}}, {{end}}{{quote $mw}}{{end}} },
{{- end }}
				},
{{- end }}
			},
		},
	)

	// Register lazy service with auto-detected dependencies
	lokstra_registry.RegisterLazyService({{quote $service.ServiceName}},
		"{{$service.ServiceName}}-factory",
		map[string]any{
{{- if $service.Dependencies }}
			"depends-on": []string{ {{range $key := sortedKeys $service.Dependencies}}{{$dep := index $service.Dependencies $key}}{{quote $dep.ServiceName}}, {{end}}},
{{- end }}
{{- range $key := sortedKeys $service.ConfigDependencies }}
{{- $cfg := index $service.ConfigDependencies $key }}
			{{quote $cfg.ConfigKey}}: lokstra_registry.GetConfig({{quote $cfg.ConfigKey}}, {{getDefaultValue $cfg.FieldType $cfg.DefaultValue}}),
{{- end }}
		})
}
{{end}}
{{end}}
{{range $filename, $code := .PreservedSections}}{{$code}}{{end}}`))

func hasReturnValue(route string) bool {
	return !strings.Contains(route, "error")
}

func extractReturnType(route string) string {
	return "*domain.User"
}
