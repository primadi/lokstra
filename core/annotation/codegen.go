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

	// Find @EndpointService and @Service annotations
	for _, ann := range file.Annotations {
		if ann.Name != "EndpointService" && ann.Name != "Service" {
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
			// @EndpointService needs name, prefix, middlewares
			args, err := ann.ReadArgs("name", "prefix", "middlewares")
			if err != nil {
				return fmt.Errorf("@EndpointService on line %d: %w", ann.Line, err)
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

		// Find @Inject annotations for dependencies and config values
		if err := extractDependencies(file, service); err != nil {
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
		// @Inject "user-repository"              - Direct service injection
		// @Inject service="user-repository"      - Direct service injection (named param)
		// @Inject "@store.implementation"        - Service name from config
		// @Inject service="@store.implementation" - Service name from config (named param)
		// @Inject "cfg:app.timeout"              - Config value injection
		// @Inject "cfg:app.timeout", "default"   - Config with default value
		// @Inject "cfg:@jwt.key-path"            - Config value via indirection
		args, err := ann.ReadArgs("service", "default")
		if err != nil {
			return fmt.Errorf("@Inject on line %d: %w", ann.Line, err)
		}

		var serviceName string
		if svc, ok := args["service"].(string); ok {
			serviceName = svc
		}

		var defaultValue string
		if def, ok := args["default"].(string); ok {
			defaultValue = def
		}

		if serviceName != "" && ann.TargetName != "" {
			// ann.TargetName is field name
			fieldType := fieldTypes[ann.TargetName]

			// Check if this is config value injection (cfg: prefix)
			if after, ok := strings.CutPrefix(serviceName, "cfg:"); ok {
				configKey := after
				// Check for indirection (@ prefix after cfg:)
				if after0, ok0 := strings.CutPrefix(configKey, "@"); ok0 {
					indirectKey := after0
					service.ConfigDependencies[configKey] = &ConfigInfo{
						ConfigKey:    configKey,
						FieldName:    ann.TargetName,
						FieldType:    fieldType,
						DefaultValue: defaultValue,
						IsIndirect:   true,
						IndirectKey:  indirectKey,
					}
				} else {
					service.ConfigDependencies[configKey] = &ConfigInfo{
						ConfigKey:    configKey,
						FieldName:    ann.TargetName,
						FieldType:    fieldType,
						DefaultValue: defaultValue,
					}
				}
			} else if after0, ok0 := strings.CutPrefix(serviceName, "@"); ok0 {
				// Config-based service injection (@ prefix)
				configKey := after0
				service.Dependencies[configKey] = &DependencyInfo{
					ServiceName:   "", // Will be resolved from config at runtime
					FieldName:     ann.TargetName,
					FieldType:     fieldType,
					IsConfigBased: true,
					ConfigKey:     configKey,
				}
			} else {
				// Direct service injection (existing behavior)
				service.Dependencies[serviceName] = &DependencyInfo{
					ServiceName:   serviceName,
					FieldName:     ann.TargetName,
					FieldType:     fieldType,
					IsConfigBased: false,
					ConfigKey:     "",
				}
			}
		}
	}

	return nil
}

// checkInitMethod checks if struct has Init() error or Init() method
func checkInitMethod(file *FileToProcess, service *ServiceGeneration) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file.FullPath, nil, parser.ParseComments)
	if err != nil {
		return
	}

	// Look for method: func (receiver *StructName) Init() error OR func (receiver *StructName) Init()
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

		// Check signature: no params
		if funcDecl.Type.Params.NumFields() != 0 {
			continue
		}

		// Check return type
		if funcDecl.Type.Results == nil || funcDecl.Type.Results.NumFields() == 0 {
			// Init() - no return value
			service.HasInitMethod = true
			service.InitReturnsError = false
			return
		}

		if funcDecl.Type.Results.NumFields() == 1 {
			// Check return type is error
			returnType := funcDecl.Type.Results.List[0].Type
			if ident, ok := returnType.(*ast.Ident); ok && ident.Name == "error" {
				// Init() error
				service.HasInitMethod = true
				service.InitReturnsError = true
				return
			}
		}
	}
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

	// Determine which hardcoded imports are actually needed
	needsDeploy := false
	needsProxy := false

	for _, service := range ctx.GeneratedCode.Services {
		if !service.IsService {
			// @EndpointService needs deploy and proxy
			needsDeploy = true
			needsProxy = true
		}
	}

	// Also check preserved sections for RouterService usage
	for _, code := range ctx.GeneratedCode.PreservedSections {
		if strings.Contains(code, "deploy.ServiceTypeConfig") || strings.Contains(code, "deploy.RouteConfig") {
			needsDeploy = true
		}
		if strings.Contains(code, "proxy.Service") || strings.Contains(code, "proxy.Call") {
			needsProxy = true
		}
	}

	// Collect used packages from method signatures, dependencies, and struct name
	usedPackages := make(map[string]bool)
	needsStrconv := false
	needsStrings := false

	for _, service := range ctx.GeneratedCode.Services {
		// From method signatures - ONLY for @EndpointService (not @Service)
		// @Service doesn't generate proxy methods, so method signatures are not in generated code
		if !service.IsService {
			for _, method := range service.Methods {
				collectPackagesFromType(method.ParamType, usedPackages)
				collectPackagesFromType(method.ReturnType, usedPackages)
			}
		}
		// From dependencies - ONLY injected fields
		for _, dep := range service.Dependencies {
			collectPackagesFromType(dep.FieldType, usedPackages)
		}
		// From config dependencies - check if time.Duration or slice types are used
		for _, cfg := range service.ConfigDependencies {
			if cfg.FieldType == "time.Duration" {
				usedPackages["time"] = true
			}
			// Check if slice types (other than []byte) are used - need strconv and strings
			if strings.HasPrefix(cfg.FieldType, "[]") && cfg.FieldType != "[]byte" {
				needsStrconv = true
				needsStrings = true
			}
			// Check if struct types are used - need cast
			if isStructType(cfg.FieldType) || isSliceOfStruct(cfg.FieldType) {
				usedPackages["cast"] = true
			}
		}
		// From struct name itself (e.g., if service struct is domain.UserService)
		collectPackagesFromType(service.StructName, usedPackages)
	}

	// Also collect packages from preserved sections (unchanged files)
	for _, code := range ctx.GeneratedCode.PreservedSections {
		extractPackagesFromCode(code, usedPackages)
	}

	// Filter imports to only used packages
	// Strategy:
	// 1. Same path + different aliases → Merge to longest alias (canonical)
	// 2. Different paths + same alias → Rename one with counter suffix

	type importEntry struct {
		Alias string
		Path  string
	}

	// Step 1: Collect all (alias, path) pairs from source files
	var allImportEntries []importEntry
	seenCombinations := make(map[string]bool) // "alias:path" -> true

	// Hardcoded imports - conditionally included based on usage
	hardcodedImports := map[string]bool{
		"github.com/primadi/lokstra/core/deploy":      needsDeploy,
		"github.com/primadi/lokstra/core/proxy":       needsProxy,
		"github.com/primadi/lokstra/lokstra_registry": true, // Always needed
	}

	// Collect from updated services
	for _, service := range ctx.GeneratedCode.Services {
		for alias, importPath := range service.Imports {
			if _, isHardcoded := hardcodedImports[importPath]; isHardcoded {
				continue
			}
			if usedPackages[alias] {
				combo := alias + ":" + importPath
				if !seenCombinations[combo] {
					allImportEntries = append(allImportEntries, importEntry{Alias: alias, Path: importPath})
					seenCombinations[combo] = true
				}
			}
		}
	}

	// Collect from existing generated file
	for alias, importPath := range existingImports {
		if _, isHardcoded := hardcodedImports[importPath]; isHardcoded {
			continue
		}
		if usedPackages[alias] {
			combo := alias + ":" + importPath
			if !seenCombinations[combo] {
				allImportEntries = append(allImportEntries, importEntry{Alias: alias, Path: importPath})
				seenCombinations[combo] = true
			}
		}
	}

	// Step 2: Group by path and find canonical alias (longest) for same path
	pathToAliases := make(map[string][]string) // path -> [aliases]
	for _, entry := range allImportEntries {
		pathToAliases[entry.Path] = append(pathToAliases[entry.Path], entry.Alias)
	}

	pathToCanonical := make(map[string]string) // path -> canonical alias (longest)

	for path, aliases := range pathToAliases {
		// Find longest alias as canonical; if tie, pick lexicographically smallest (alphabetically first)
		canonical := aliases[0]
		for i, alias := range aliases {
			if i == 0 {
				continue
			}
			if len(alias) > len(canonical) || (len(alias) == len(canonical) && alias < canonical) {
				canonical = alias
			}
		}
		pathToCanonical[path] = canonical
	}

	// Step 3: Build final import list with conflict resolution
	aliasToPath := make(map[string]string)      // alias -> path (for conflict detection)
	pathToFinalAlias := make(map[string]string) // path -> final alias (after conflict resolution)
	var finalImports []importEntry

	for path, canonical := range pathToCanonical {
		// Check if canonical alias conflicts with another path
		if existingPath, exists := aliasToPath[canonical]; exists && existingPath != path {
			// Conflict! Rename this one
			newAlias := canonical
			counter := 1
			for {
				newAlias = fmt.Sprintf("%s_%d", canonical, counter)
				if _, taken := aliasToPath[newAlias]; !taken {
					break
				}
				counter++
			}
			finalImports = append(finalImports, importEntry{Alias: newAlias, Path: path})
			aliasToPath[newAlias] = path
			pathToFinalAlias[path] = newAlias
		} else {
			finalImports = append(finalImports, importEntry{Alias: canonical, Path: path})
			aliasToPath[canonical] = path
			pathToFinalAlias[path] = canonical
		}
	}

	// Step 3.5: Build alias remap based on actual final aliases
	// Map: (path, oldAlias) -> finalAlias
	aliasRemap := make(map[string]map[string]string) // path -> (oldAlias -> finalAlias)

	for path, aliases := range pathToAliases {
		finalAlias := pathToFinalAlias[path]
		aliasRemap[path] = make(map[string]string)

		for _, oldAlias := range aliases {
			aliasRemap[path][oldAlias] = finalAlias
		}
	}

	// Step 4: Add standard library imports if needed
	if usedPackages["time"] {
		if _, exists := aliasToPath["time"]; !exists {
			finalImports = append(finalImports, importEntry{Alias: "time", Path: "time"})
			aliasToPath["time"] = "time"
		}
	}

	if needsStrconv {
		if _, exists := aliasToPath["strconv"]; !exists {
			finalImports = append(finalImports, importEntry{Alias: "strconv", Path: "strconv"})
			aliasToPath["strconv"] = "strconv"
		}
	}

	if needsStrings {
		if _, exists := aliasToPath["strings"]; !exists {
			finalImports = append(finalImports, importEntry{Alias: "strings", Path: "strings"})
			aliasToPath["strings"] = "strings"
		}
	}

	if usedPackages["cast"] {
		if _, exists := aliasToPath["cast"]; !exists {
			finalImports = append(finalImports, importEntry{Alias: "cast", Path: "github.com/primadi/lokstra/common/cast"})
			aliasToPath["cast"] = "github.com/primadi/lokstra/common/cast"
		}
	}

	// Step 5: Apply alias remapping to all services
	// Update method signatures and dependency types to use final aliases
	for _, service := range ctx.GeneratedCode.Services {
		// Build remap for this service based on its imports
		serviceRemap := make(map[string]string)
		for oldAlias, path := range service.Imports {
			if pathRemap, exists := aliasRemap[path]; exists {
				if finalAlias, exists := pathRemap[oldAlias]; exists && finalAlias != oldAlias {
					serviceRemap[oldAlias] = finalAlias
				}
			}
		}

		// Remap method signatures
		for _, method := range service.Methods {
			if method.ParamType != "" {
				method.ParamType = remapTypeAliases(method.ParamType, serviceRemap)
			}
			if method.ReturnType != "" {
				method.ReturnType = remapTypeAliases(method.ReturnType, serviceRemap)
			}
		}

		// Remap dependency types
		for _, dep := range service.Dependencies {
			if dep.FieldType != "" {
				dep.FieldType = remapTypeAliases(dep.FieldType, serviceRemap)
			}
		}

		// Remap config dependency types
		for _, cfg := range service.ConfigDependencies {
			if cfg.FieldType != "" {
				cfg.FieldType = remapTypeAliases(cfg.FieldType, serviceRemap)
			}
		}
	}

	// Generate code
	var buf bytes.Buffer
	if err := genTemplate.Execute(&buf, map[string]any{
		"Package":           pkgName,
		"Services":          ctx.GeneratedCode.Services,
		"PreservedSections": ctx.GeneratedCode.PreservedSections,
		"AllImports":        finalImports,
		"AllStructNames":    allStructNames,
		"NeedsDeploy":       needsDeploy,
		"NeedsProxy":        needsProxy,
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

	// Remove pointer and array prefixes FIRST before checking for generics
	// This handles: []*domain.User, *[]domain.User, etc.
	cleanType := strings.TrimLeft(typeStr, "*[]")

	// Handle generics: Type[Param1, Param2]
	// Generic brackets appear AFTER the type name, not at the start
	if start := strings.Index(cleanType, "["); start != -1 {
		if end := strings.LastIndex(cleanType, "]"); end > start {
			// Process the base type (before '[')
			baseType := cleanType[:start]
			collectPackagesFromType(baseType, packages)

			// Process inner type parameters
			innerTypes := cleanType[start+1 : end]
			// Split by comma for multiple type params
			for _, inner := range strings.Split(innerTypes, ",") {
				collectPackagesFromType(strings.TrimSpace(inner), packages)
			}
			return
		}
	}

	// Extract package prefix (everything before last dot)
	if idx := strings.LastIndex(cleanType, "."); idx != -1 {
		pkg := cleanType[:idx]
		// Handle nested packages (e.g., "github.com/user/repo.Type" -> "repo")
		// Only take the last segment as the alias
		if lastSlash := strings.LastIndex(pkg, "/"); lastSlash != -1 {
			pkg = pkg[lastSlash+1:]
		}
		packages[pkg] = true
	}
}

// remapTypeAliases remaps package aliases in type strings
// E.g., "models.User" -> "pkgamodel.User" if aliasMap["models"] = "pkgamodel"
// Handles: pkg.Type, *pkg.Type, []pkg.Type, []*pkg.Type, map[pkg.Key]pkg.Value, etc.
func remapTypeAliases(typeStr string, aliasMap map[string]string) string {
	if typeStr == "" || len(aliasMap) == 0 {
		return typeStr
	}

	// Replace all qualified identifiers: pkg.Type
	// Pattern: word boundary + package name + dot + identifier
	for oldAlias, newAlias := range aliasMap {
		if oldAlias == newAlias {
			continue // No change needed
		}

		// Use regex to match: (^|[^a-zA-Z0-9_])oldAlias\.
		// This ensures we match "models.User" but not "mymodels.User"
		pattern := `(^|[^a-zA-Z0-9_])` + regexp.QuoteMeta(oldAlias) + `\.`
		re := regexp.MustCompile(pattern)

		// Replace with: $1newAlias.
		typeStr = re.ReplaceAllString(typeStr, "${1}"+newAlias+".")
	}

	return typeStr
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

// parseIndirectConfigValue generates code for indirect config resolution
// First resolves the actual config key from another config value, then fetches the final value
func parseIndirectConfigValue(fieldType, indirectKey, defaultValue string) string {
	// Generate appropriate default value based on field type
	defaultVal := defaultValue
	if defaultVal == "" {
		switch fieldType {
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			defaultVal = "0"
		case "bool":
			defaultVal = "false"
		case "float32", "float64":
			defaultVal = "0.0"
		case "time.Duration":
			defaultVal = "0"
		case "[]byte":
			defaultVal = "nil"
		case "string":
			defaultVal = `""`
		default:
			if strings.HasPrefix(fieldType, "[]") {
				defaultVal = "nil"
			} else if strings.HasPrefix(fieldType, "*") {
				defaultVal = "nil"
			} else if isStructType(fieldType) {
				defaultVal = fieldType + "{}"
			} else {
				defaultVal = `""`
			}
		}
	}

	// Generate code for indirect resolution
	// Step 1: Get actual key name from config[indirectKey]
	// Step 2: Get actual value from config[actualKey]
	// Step 3: Parse based on field type
	return generateIndirectConfigCode(fieldType, indirectKey, defaultVal)
}

// generateIndirectConfigCode generates the actual Go code for indirect config resolution
func generateIndirectConfigCode(fieldType, indirectKey, defaultVal string) string {
	// Different handling based on field type
	if fieldType == "string" {
		return fmt.Sprintf(`func() string {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if v, ok := cfg[actualKey].(string); ok { return v }
		}
		return %s
	}()`, indirectKey, defaultVal)
	}

	if fieldType == "int" || fieldType == "int64" {
		return fmt.Sprintf(`func() %s {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if v, ok := cfg[actualKey].(%s); ok { return v }
			if v, ok := cfg[actualKey].(float64); ok { return %s(v) }
		}
		return %s
	}()`, fieldType, indirectKey, fieldType, fieldType, defaultVal)
	}

	if fieldType == "float64" {
		return fmt.Sprintf(`func() float64 {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if v, ok := cfg[actualKey].(float64); ok { return v }
			if v, ok := cfg[actualKey].(int); ok { return float64(v) }
		}
		return %s
	}()`, indirectKey, defaultVal)
	}

	if fieldType == "bool" {
		return fmt.Sprintf(`func() bool {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if v, ok := cfg[actualKey].(bool); ok { return v }
		}
		return %s
	}()`, indirectKey, defaultVal)
	}

	if fieldType == "time.Duration" {
		return fmt.Sprintf(`func() time.Duration {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if v, ok := cfg[actualKey].(time.Duration); ok { return v }
			if s, ok := cfg[actualKey].(string); ok {
				if d, err := time.ParseDuration(s); err == nil { return d }
			}
		}
		return %s
	}()`, indirectKey, defaultVal)
	}

	if fieldType == "[]byte" {
		return fmt.Sprintf(`func() []byte {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if v, ok := cfg[actualKey].([]byte); ok { return v }
			if s, ok := cfg[actualKey].(string); ok { return []byte(s) }
		}
		return %s
	}()`, indirectKey, defaultVal)
	}

	// For slices
	if strings.HasPrefix(fieldType, "[]") {
		elementType := strings.TrimPrefix(fieldType, "[]")
		if isStructType(elementType) {
			return fmt.Sprintf(`func() %s {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if arr, ok := cfg[actualKey].([]any); ok {
				result := make(%s, 0, len(arr))
				for _, item := range arr {
					var elem %s
					if err := cast.ToStruct(item, &elem, false); err == nil {
						result = append(result, elem)
					}
				}
				return result
			}
			if arr, ok := cfg[actualKey].(%s); ok { return arr }
		}
		return %s
	}()`, fieldType, indirectKey, fieldType, elementType, fieldType, defaultVal)
		}
		// Primitive slices
		return fmt.Sprintf(`func() %s {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if v, ok := cfg[actualKey].(%s); ok { return v }
		}
		return %s
	}()`, fieldType, indirectKey, fieldType, defaultVal)
	}

	// For struct types
	if isStructType(fieldType) {
		return fmt.Sprintf(`func() %s {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if v, ok := cfg[actualKey]; ok {
				var result %s
				if err := cast.ToStruct(v, &result, false); err == nil {
					return result
				}
			}
		}
		return %s
	}()`, fieldType, indirectKey, fieldType, defaultVal)
	}

	// Fallback
	return fmt.Sprintf(`func() %s {
		if actualKey, ok := cfg[%q].(string); ok && actualKey != "" {
			if v, ok := cfg[actualKey].(%s); ok { return v }
		}
		return %s
	}()`, fieldType, indirectKey, fieldType, defaultVal)
}

// parseDurationFromConfig generates code to parse time.Duration from config value
// Handles both string ("15m", "20h") and time.Duration values from config
// parseConfigValue generates code to parse any config value with type conversion
// This universal function handles: primitives, time.Duration, []byte, slices, structs
// configKey can start with @ for indirection (e.g., "@jwt.key-path" resolves actual key from config)
func parseConfigValue(fieldType, configKey, defaultValue string) string {
	// Check if this is indirect config resolution (@ prefix)
	if strings.HasPrefix(configKey, "@") {
		indirectKey := strings.TrimPrefix(configKey, "@")
		return parseIndirectConfigValue(fieldType, indirectKey, defaultValue)
	}
	// For primitives (string, int, bool, float64), use direct type assertion
	if fieldType == "string" || fieldType == "int" || fieldType == "int64" ||
		fieldType == "float64" || fieldType == "bool" {
		if defaultValue != "" {
			return fmt.Sprintf("cfg[%q].(%s)", configKey, fieldType)
		}
		return fmt.Sprintf("cfg[%q].(%s)", configKey, fieldType)
	}

	// For time.Duration - handle both duration and string
	if fieldType == "time.Duration" {
		defaultVal := defaultValue
		if defaultVal == "" {
			defaultVal = "0"
		}
		return fmt.Sprintf(`func() time.Duration {
		if v, ok := cfg[%q].(time.Duration); ok { return v }
		if s, ok := cfg[%q].(string); ok {
			if d, err := time.ParseDuration(s); err == nil { return d }
		}
		return %s
	}()`, configKey, configKey, defaultVal)
	}

	// For []byte - handle both []byte and string
	if fieldType == "[]byte" {
		defaultVal := "nil"
		if defaultValue != "" {
			defaultVal = "[]byte(" + defaultValue + ")"
		}
		return fmt.Sprintf(`func() []byte {
		if v, ok := cfg[%q].([]byte); ok { return v }
		if s, ok := cfg[%q].(string); ok { return []byte(s) }
		return %s
	}()`, configKey, configKey, defaultVal)
	}

	// For slices ([]string, []int, []struct, etc.)
	if strings.HasPrefix(fieldType, "[]") {
		elementType := strings.TrimPrefix(fieldType, "[]")
		defaultVal := "nil"
		if defaultValue != "" {
			defaultVal = defaultValue
		}

		// For slice of struct, use cast.ToStruct per element
		if isStructType(elementType) {
			return fmt.Sprintf(`func() %s {
		if arr, ok := cfg[%q].([]any); ok {
			result := make(%s, 0, len(arr))
			for _, item := range arr {
				var elem %s
				if err := cast.ToStruct(item, &elem, false); err == nil {
					result = append(result, elem)
				}
			}
			return result
		}
		if arr, ok := cfg[%q].(%s); ok { return arr }
		return %s
	}()`, fieldType, configKey, fieldType, elementType, configKey, fieldType, defaultVal)
		}

		// For primitive slices ([]string, []int, etc.) - handle array and comma-separated string
		var parseCode string
		switch elementType {
		case "string":
			parseCode = `if arr, ok := cfg[%q].([]any); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok { result = append(result, s) }
			}
			return result
		}
		if arr, ok := cfg[%q].([]string); ok { return arr }
		if s, ok := cfg[%q].(string); ok {
			if s != "" {
				parts := strings.Split(s, ",")
				result := make([]string, 0, len(parts))
				for _, p := range parts {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						result = append(result, trimmed)
					}
				}
				return result
			}
		}`
		case "int":
			parseCode = `if arr, ok := cfg[%q].([]any); ok {
			result := make([]int, 0, len(arr))
			for _, item := range arr {
				switch v := item.(type) {
				case int: result = append(result, v)
				case float64: result = append(result, int(v))
				}
			}
			return result
		}
		if arr, ok := cfg[%q].([]int); ok { return arr }
		if s, ok := cfg[%q].(string); ok {
			if s != "" {
				parts := strings.Split(s, ",")
				result := make([]int, 0, len(parts))
				for _, p := range parts {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						if val, err := strconv.Atoi(trimmed); err == nil {
							result = append(result, val)
						}
					}
				}
				return result
			}
		}`
		case "int64":
			parseCode = `if arr, ok := cfg[%q].([]any); ok {
			result := make([]int64, 0, len(arr))
			for _, item := range arr {
				switch v := item.(type) {
				case int64: result = append(result, v)
				case int: result = append(result, int64(v))
				case float64: result = append(result, int64(v))
				}
			}
			return result
		}
		if arr, ok := cfg[%q].([]int64); ok { return arr }
		if s, ok := cfg[%q].(string); ok {
			if s != "" {
				parts := strings.Split(s, ",")
				result := make([]int64, 0, len(parts))
				for _, p := range parts {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						if val, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
							result = append(result, val)
						}
					}
				}
				return result
			}
		}`
		case "float64":
			parseCode = `if arr, ok := cfg[%q].([]any); ok {
			result := make([]float64, 0, len(arr))
			for _, item := range arr {
				switch v := item.(type) {
				case float64: result = append(result, v)
				case int: result = append(result, float64(v))
				case int64: result = append(result, float64(v))
				}
			}
			return result
		}
		if arr, ok := cfg[%q].([]float64); ok { return arr }
		if s, ok := cfg[%q].(string); ok {
			if s != "" {
				parts := strings.Split(s, ",")
				result := make([]float64, 0, len(parts))
				for _, p := range parts {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						if val, err := strconv.ParseFloat(trimmed, 64); err == nil {
							result = append(result, val)
						}
					}
				}
				return result
			}
		}`
		default:
			// Generic slice handling
			parseCode = fmt.Sprintf(`if arr, ok := cfg[%%q].(%s); ok { return arr }`, fieldType)
		}

		return fmt.Sprintf(`func() %s {
		%s
		return %s
	}()`, fieldType, fmt.Sprintf(parseCode, configKey, configKey, configKey), defaultVal)
	}

	// For struct types - use cast.ToStruct
	if isStructType(fieldType) {
		defaultVal := fieldType + "{}"
		if defaultValue != "" {
			defaultVal = defaultValue
		}
		return fmt.Sprintf(`func() %s {
		if v, ok := cfg[%q]; ok {
			var result %s
			if err := cast.ToStruct(v, &result, false); err == nil {
				return result
			}
		}
		return %s
	}()`, fieldType, configKey, fieldType, defaultVal)
	}

	// Fallback: direct type assertion
	return fmt.Sprintf("cfg[%q].(%s)", configKey, fieldType)
}

func parseDurationFromConfig(configKey, defaultValue string) string {
	if defaultValue != "" {
		return fmt.Sprintf(`func() time.Duration {
		if v, ok := cfg[%q].(time.Duration); ok {
			return v
		}
		if s, ok := cfg[%q].(string); ok {
			if d, err := time.ParseDuration(s); err == nil {
				return d
			}
		}
		return %s
	}()`, configKey, configKey, defaultValue)
	}
	return fmt.Sprintf(`func() time.Duration {
		if v, ok := cfg[%q].(time.Duration); ok {
			return v
		}
		if s, ok := cfg[%q].(string); ok {
			if d, err := time.ParseDuration(s); err == nil {
				return d
			}
		}
		return 0
	}()`, configKey, configKey)
}

// parseByteSliceFromConfig generates code to parse []byte from config value
// Handles both []byte and string values from config
func parseByteSliceFromConfig(configKey, defaultValue string) string {
	if defaultValue != "" {
		return fmt.Sprintf(`func() []byte {
		if v, ok := cfg[%q].([]byte); ok {
			return v
		}
		if s, ok := cfg[%q].(string); ok {
			return []byte(s)
		}
		return []byte(%s)
	}()`, configKey, configKey, defaultValue)
	}
	return fmt.Sprintf(`func() []byte {
		if v, ok := cfg[%q].([]byte); ok {
			return v
		}
		if s, ok := cfg[%q].(string); ok {
			return []byte(s)
		}
		return nil
	}()`, configKey, configKey)
}

// parseSliceFromConfig generates code to parse slice types from config value
// Handles []string, []int, []int64, []float64, etc.
// Supports both array values and comma-separated strings
// parseStructFromConfig generates code to parse struct from config using cast.ToStruct
func parseStructFromConfig(fieldType, configKey, defaultValue string) string {
	if defaultValue != "" {
		return fmt.Sprintf(`func() %s {
		if v, ok := cfg[%q]; ok {
			var result %s
			if err := cast.ToStruct(v, &result, false); err == nil {
				return result
			}
		}
		return %s
	}()`, fieldType, configKey, fieldType, defaultValue)
	}
	return fmt.Sprintf(`func() %s {
		if v, ok := cfg[%q]; ok {
			var result %s
			if err := cast.ToStruct(v, &result, false); err == nil {
				return result
			}
		}
		return %s{}
	}()`, fieldType, configKey, fieldType, fieldType)
}

func parseSliceFromConfig(fieldType, configKey, defaultValue string) string {
	// Extract element type from []ElementType
	elementType := strings.TrimPrefix(fieldType, "[]")

	var conversionCode string
	switch elementType {
	case "string":
		conversionCode = `if arr, ok := cfg[%q].([]any); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
		if arr, ok := cfg[%q].([]string); ok {
			return arr
		}
		if s, ok := cfg[%q].(string); ok {
			if s != "" {
				parts := strings.Split(s, ",")
				result := make([]string, 0, len(parts))
				for _, p := range parts {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						result = append(result, trimmed)
					}
				}
				return result
			}
		}`
	case "int":
		conversionCode = `if arr, ok := cfg[%q].([]any); ok {
			result := make([]int, 0, len(arr))
			for _, item := range arr {
				switch v := item.(type) {
				case int:
					result = append(result, v)
				case float64:
					result = append(result, int(v))
				}
			}
			return result
		}
		if arr, ok := cfg[%q].([]int); ok {
			return arr
		}
		if s, ok := cfg[%q].(string); ok {
			if s != "" {
				parts := strings.Split(s, ",")
				result := make([]int, 0, len(parts))
				for _, p := range parts {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						if val, err := strconv.Atoi(trimmed); err == nil {
							result = append(result, val)
						}
					}
				}
				return result
			}
		}`
	case "int64":
		conversionCode = `if arr, ok := cfg[%q].([]any); ok {
			result := make([]int64, 0, len(arr))
			for _, item := range arr {
				switch v := item.(type) {
				case int64:
					result = append(result, v)
				case int:
					result = append(result, int64(v))
				case float64:
					result = append(result, int64(v))
				}
			}
			return result
		}
		if arr, ok := cfg[%q].([]int64); ok {
			return arr
		}
		if s, ok := cfg[%q].(string); ok {
			if s != "" {
				parts := strings.Split(s, ",")
				result := make([]int64, 0, len(parts))
				for _, p := range parts {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						if val, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
							result = append(result, val)
						}
					}
				}
				return result
			}
		}`
	case "float64":
		conversionCode = `if arr, ok := cfg[%q].([]any); ok {
			result := make([]float64, 0, len(arr))
			for _, item := range arr {
				switch v := item.(type) {
				case float64:
					result = append(result, v)
				case int:
					result = append(result, float64(v))
				case int64:
					result = append(result, float64(v))
				}
			}
			return result
		}
		if arr, ok := cfg[%q].([]float64); ok {
			return arr
		}
		if s, ok := cfg[%q].(string); ok {
			if s != "" {
				parts := strings.Split(s, ",")
				result := make([]float64, 0, len(parts))
				for _, p := range parts {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						if val, err := strconv.ParseFloat(trimmed, 64); err == nil {
							result = append(result, val)
						}
					}
				}
				return result
			}
		}`
	default:
		// For struct types, use cast.ToStruct for conversion
		if isStructType(elementType) {
			conversionCode = fmt.Sprintf(`if arr, ok := cfg[%%q].([]any); ok {
			result := make(%s, 0, len(arr))
			for _, item := range arr {
				var elem %s
				if err := cast.ToStruct(item, &elem, false); err == nil {
					result = append(result, elem)
				}
			}
			return result
		}
		if arr, ok := cfg[%%q].(%s); ok {
			return arr
		}`, fieldType, elementType, fieldType)
		} else {
			// For other slice types ([]any, []interface{}, custom types), try direct type assertion
			conversionCode = fmt.Sprintf(`if arr, ok := cfg[%%q].(%s); ok {
			return arr
		}
		if arr, ok := cfg[%%q].([]any); ok {
			result := make(%s, 0, len(arr))
			for _, item := range arr {
				if v, ok := item.(%s); ok {
					result = append(result, v)
				}
			}
			return result
		}`, fieldType, fieldType, elementType)
		}
	}

	if defaultValue != "" {
		return fmt.Sprintf(`func() %s {
		%s
		return %s
	}()`, fieldType, fmt.Sprintf(conversionCode, configKey, configKey, configKey), defaultValue)
	}
	return fmt.Sprintf(`func() %s {
		%s
		return nil
	}()`, fieldType, fmt.Sprintf(conversionCode, configKey, configKey, configKey))
}

// genTemplate is the template for zz_generated.lokstra.go
var genTemplate = template.Must(template.New("gen").Funcs(template.FuncMap{
	"quote":      func(s string) string { return fmt.Sprintf("%q", s) },
	"join":       strings.Join,
	"trimPrefix": strings.TrimPrefix,
	"trimSuffix": strings.TrimSuffix,
	"notEmpty":   func(s string) bool { return strings.TrimSpace(s) != "" },

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
			}

			// For slice types, check if it's a Go slice literal or convert from string
			if strings.HasPrefix(fieldType, "[]") {
				// If already a slice literal (starts with fieldType{ or []...{), use as-is
				if strings.HasPrefix(defaultValue, fieldType+"{") || strings.HasPrefix(defaultValue, "[]") {
					return defaultValue
				}
				// Otherwise, try to parse as simplified format based on element type
				elementType := strings.TrimPrefix(fieldType, "[]")
				switch elementType {
				case "string":
					// Parse "a,b,c" -> []string{"a", "b", "c"}
					parts := strings.Split(defaultValue, ",")
					quotedParts := make([]string, 0, len(parts))
					for _, p := range parts {
						trimmed := strings.TrimSpace(p)
						if trimmed != "" {
							quotedParts = append(quotedParts, fmt.Sprintf(`"%s"`, trimmed))
						}
					}
					return fmt.Sprintf("%s{%s}", fieldType, strings.Join(quotedParts, ", "))
				case "int", "int64", "float64":
					// Parse "1,2,3" -> []int{1, 2, 3}
					parts := strings.Split(defaultValue, ",")
					trimmedParts := make([]string, 0, len(parts))
					for _, p := range parts {
						if trimmed := strings.TrimSpace(p); trimmed != "" {
							trimmedParts = append(trimmedParts, trimmed)
						}
					}
					return fmt.Sprintf("%s{%s}", fieldType, strings.Join(trimmedParts, ", "))
				default:
					// For other types, assume it's already a valid Go literal
					return defaultValue
				}
			}

			// For int/bool/float, return as-is
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
		case "[]byte":
			return "nil"
		case "string":
			return `""`
		default:
			// Check if it's a slice type
			if strings.HasPrefix(fieldType, "[]") {
				return "nil"
			}
			// Check if it's a pointer
			if strings.HasPrefix(fieldType, "*") {
				return "nil"
			}
			// For struct types, return zero value: StructType{}
			if isStructType(fieldType) {
				return fieldType + "{}"
			}
			// Fallback for unknown types
			return `""`
		}
	},
	"parseDurationFromConfig":  parseDurationFromConfig,
	"parseByteSliceFromConfig": parseByteSliceFromConfig,
	"parseSliceFromConfig":     parseSliceFromConfig,
	"parseStructFromConfig":    parseStructFromConfig,
	"parseConfigValue":         parseConfigValue, // Universal config parser
	"isSliceType":              func(fieldType string) bool { return strings.HasPrefix(fieldType, "[]") && fieldType != "[]byte" },
	"isStructType":             isStructType,
	"isSliceOfStruct":          isSliceOfStruct,
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
// Annotations: @EndpointService, @Service, @Inject, @Route
//
// @Inject supports:
//   - "service-name"          : Direct service injection
//   - "@config.key"           : Service name from config value
//   - "cfg:config.key"        : Config value injection
//   - "cfg:@config.key"       : Config value via indirection

package {{.Package}}

import (
{{- if .NeedsDeploy }}
	"github.com/primadi/lokstra/core/deploy"
{{- end }}
{{- if .NeedsProxy }}
	"github.com/primadi/lokstra/core/proxy"
{{- end }}
	"github.com/primadi/lokstra/lokstra_registry"
{{- range $entry := .AllImports }}
	{{$entry.Alias}} "{{$entry.Path}}"
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
{{- if or $service.Dependencies $service.ConfigDependencies}}
//   - @Inject annotations
{{- end}}
func Register{{$service.StructName}}() {
	lokstra_registry.RegisterLazyService({{quote $service.ServiceName}}, func(deps map[string]any, cfg map[string]any) any {
		svc := &{{$service.StructName}}{
{{- range $key := sortedKeys $service.Dependencies }}
{{- $dep := index $service.Dependencies $key }}
{{- if $dep.IsConfigBased }}
			{{$dep.FieldName}}: deps["@{{$dep.ConfigKey}}"].({{$dep.FieldType}}),
{{- else }}
			{{$dep.FieldName}}: deps[{{quote $dep.ServiceName}}].({{$dep.FieldType}}),
{{- end }}
{{- end }}
{{- range $key := sortedKeys $service.ConfigDependencies }}
{{- $cfg := index $service.ConfigDependencies $key }}
			{{$cfg.FieldName}}: {{parseConfigValue $cfg.FieldType $cfg.ConfigKey (getDefaultValue $cfg.FieldType $cfg.DefaultValue)}},
{{- end }}
		}
{{- if $service.HasInitMethod }}
		
		// Call Init() for post-initialization
{{- if $service.InitReturnsError }}
		if err := svc.Init(); err != nil {
			panic("failed to initialize {{$service.ServiceName}}: " + err.Error())
		}
{{- else }}
		svc.Init()
{{- end }}
{{- end }}
		
		return svc
	}, map[string]any{
{{- if or $service.Dependencies $service.ConfigDependencies }}
{{- if $service.Dependencies }}
		"depends-on": []string{ {{range $key := sortedKeys $service.Dependencies}}{{$dep := index $service.Dependencies $key}}{{if $dep.IsConfigBased}}"@{{$dep.ConfigKey}}", {{else}}{{quote $dep.ServiceName}}, {{end}}{{end}}},
{{- end }}
{{- range $key := sortedKeys $service.ConfigDependencies }}
{{- $cfg := index $service.ConfigDependencies $key }}
		{{quote $cfg.ConfigKey}}: lokstra_registry.GetConfig[any]({{quote $cfg.ConfigKey}}, {{getDefaultValue $cfg.FieldType $cfg.DefaultValue}}),
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
{{- if $dep.IsConfigBased }}
		{{$dep.FieldName}}: deps["@{{$dep.ConfigKey}}"].({{$dep.FieldType}}),
{{- else }}
		{{$dep.FieldName}}: deps[{{quote $dep.ServiceName}}].({{$dep.FieldType}}),
{{- end }}
{{- end }}
{{- range $key := sortedKeys $service.ConfigDependencies }}
{{- $cfg := index $service.ConfigDependencies $key }}
		{{$cfg.FieldName}}: {{parseConfigValue $cfg.FieldType $cfg.ConfigKey (getDefaultValue $cfg.FieldType $cfg.DefaultValue)}},
{{- end }}
	}
{{- if $service.HasInitMethod }}
	
	// Call Init() for post-initialization
{{- if $service.InitReturnsError }}
	if err := svc.Init(); err != nil {
		panic("failed to initialize {{$service.ServiceName}}: " + err.Error())
	}
{{- else }}
	svc.Init()
{{- end }}
{{- end }}
	
	return svc
}

// {{$service.RemoteTypeName}}Factory creates a remote HTTP client for {{$service.InterfaceName}}
// Auto-generated from @EndpointService annotation
func {{$service.RemoteTypeName}}Factory(deps, config map[string]any) any {
	proxyService, ok := config["remote"].(*proxy.Service)
	if !ok {
		panic("remote factory requires 'remote' (proxy.Service) in config")
	}
	return New{{$service.RemoteTypeName}}(proxyService)
}

// Register{{$service.StructName}} registers the {{$service.ServiceName}} with the registry
// Auto-generated from annotations:
//   - @EndpointService name={{quote $service.ServiceName}}, prefix={{quote $service.Prefix}}
{{- if or $service.Dependencies $service.ConfigDependencies}}
//   - @Inject annotations
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
			"depends-on": []string{ {{range $key := sortedKeys $service.Dependencies}}{{$dep := index $service.Dependencies $key}}{{if $dep.IsConfigBased}}"@{{$dep.ConfigKey}}", {{else}}{{quote $dep.ServiceName}}, {{end}}{{end}}},
{{- end }}
{{- if $service.ConfigDependencies }}
{{- range $key := sortedKeys $service.ConfigDependencies }}
{{- $cfg := index $service.ConfigDependencies $key }}
			{{quote $cfg.ConfigKey}}: lokstra_registry.GetConfig[any]({{quote $cfg.ConfigKey}}, {{getDefaultValue $cfg.FieldType $cfg.DefaultValue}}),
{{- end }}
{{- end }}
		})
}
{{end}}
{{end}}
{{range $filename, $code := .PreservedSections}}{{$code}}{{end}}`))

// isStructType checks if a type is a struct (not primitive, slice, map, or interface)
func isStructType(fieldType string) bool {
	// Empty type or primitive types
	if fieldType == "" || fieldType == "string" || fieldType == "int" || fieldType == "int64" ||
		fieldType == "float64" || fieldType == "bool" || fieldType == "time.Duration" ||
		fieldType == "[]byte" {
		return false
	}
	// Slice, map, or interface types
	if strings.HasPrefix(fieldType, "[]") || strings.HasPrefix(fieldType, "map[") ||
		fieldType == "any" || fieldType == "interface{}" {
		return false
	}
	// Everything else is considered a struct
	return true
}

// isSliceOfStruct checks if a type is a slice of struct
func isSliceOfStruct(fieldType string) bool {
	if !strings.HasPrefix(fieldType, "[]") {
		return false
	}
	elementType := strings.TrimPrefix(fieldType, "[]")
	return isStructType(elementType)
}
