package annotation

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const generatedFileName = "zz_generated.lokstra.go"
const cacheFileName = "zz_cache.lokstra.json"

// GenerateCodeForFolder generates zz_generated.lokstra.go based on RouterServiceContext
func GenerateCodeForFolder(ctx *RouterServiceContext) error {
	// If all files are skipped and no files were deleted, skip generation
	if len(ctx.UpdatedFiles) == 0 && len(ctx.DeletedFiles) == 0 {
		// Nothing changed, no need to regenerate
		return nil
	}

	// Read existing zz_generated.lokstra.go to preserve code for skipped files
	genPath := filepath.Join(ctx.FolderPath, generatedFileName)
	existingGenCode := readExistingGenCode(genPath)

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

	// Generate zz_generated.lokstra.go
	if len(ctx.GeneratedCode.Services) > 0 || len(ctx.GeneratedCode.PreservedSections) > 0 {
		if err := writeGenFile(genPath, ctx); err != nil {
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

	// Find @RouterService annotations
	for _, ann := range file.Annotations {
		if ann.Name != "RouterService" {
			continue
		}

		// Read RouterService args
		args, err := ann.readArgs("name", "prefix", "middlewares")
		if err != nil {
			return fmt.Errorf("@RouterService on line %d: %w", ann.Line, err)
		}

		serviceName, _ := args["name"].(string)
		prefix, _ := args["prefix"].(string)
		middlewares := extractStringArray(args["middlewares"])

		if serviceName == "" {
			return fmt.Errorf("@RouterService on line %d: 'name' is required", ann.Line)
		}

		// Create service generation entry
		service := &ServiceGeneration{
			ServiceName:  serviceName,
			Prefix:       prefix,
			Middlewares:  middlewares,
			Routes:       make(map[string]string),
			Methods:      make(map[string]*MethodSignature),
			Dependencies: make(map[string]*DependencyInfo),
			Imports:      make(map[string]string),
			StructName:   ann.TargetName,
			SourceFile:   file.Filename,
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

		ctx.GeneratedCode.Services[serviceName] = service
	}

	return nil
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

		if recvName != structName {
			continue
		}

		// Extract method signature
		methodName := funcDecl.Name.Name
		sig := &MethodSignature{
			Name: methodName,
		}

		// Extract parameter type (assume single param)
		if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
			sig.ParamType = exprToString(funcDecl.Type.Params.List[0].Type)
		}

		// Extract return type
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
		return "interface{}"
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
		return "interface{}"
	}
}

// extractRoutes finds all @Route annotations on methods
func extractRoutes(file *FileToProcess, service *ServiceGeneration) error {
	for _, ann := range file.Annotations {
		if ann.Name != "Route" && ann.Name != "route" {
			continue
		}

		// @Route "GET /users/{id}"
		// or @Route method="GET", path="/users/{id}"
		args, err := ann.readArgs("route")
		if err != nil {
			// Try named args
			args, err = ann.readArgs("method", "path")
			if err != nil {
				return fmt.Errorf("@Route on line %d: %w", ann.Line, err)
			}
		}

		var routeStr string
		if route, ok := args["route"].(string); ok {
			routeStr = route
		} else if method, ok := args["method"].(string); ok {
			path, _ := args["path"].(string)
			routeStr = method + " " + path
		}

		if routeStr != "" && ann.TargetName != "" {
			service.Routes[ann.TargetName] = routeStr
		}
	}

	return nil
}

// extractDependencies finds all @Inject annotations and field info
func extractDependencies(file *FileToProcess, service *ServiceGeneration) error {
	// Parse AST to get field types
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file.FullPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Find struct fields
	fieldTypes := make(map[string]string) // fieldName -> fieldType
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

			// Extract field names and types
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue
				}
				fieldName := field.Names[0].Name
				fieldType := exprToString(field.Type)
				fieldTypes[fieldName] = fieldType
			}
		}
	}

	// Now process @Inject annotations
	for _, ann := range file.Annotations {
		if ann.Name != "Inject" && ann.Name != "inject" {
			continue
		}

		// @Inject "user-repository"
		// or @Inject service="user-repository"
		args, err := ann.readArgs("service")
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

			// Extract inner type from *service.Cached[T] or similar
			innerType := extractInnerGenericType(fieldType)

			service.Dependencies[serviceName] = &DependencyInfo{
				ServiceName: serviceName,
				FieldName:   ann.TargetName,
				FieldType:   fieldType,
				InnerType:   innerType,
			}
		}
	}

	return nil
}

// extractInnerGenericType extracts T from *service.Cached[T]
func extractInnerGenericType(fieldType string) string {
	// Find [ and ]
	start := strings.Index(fieldType, "[")
	end := strings.LastIndex(fieldType, "]")

	if start != -1 && end != -1 && end > start {
		return fieldType[start+1 : end]
	}

	// If not generic, return as-is
	return fieldType
}

// writeGenFile writes the zz_generated.lokstra.go file
func writeGenFile(path string, ctx *RouterServiceContext) error {
	// Get package name from existing files
	pkgName, err := getPackageName(ctx.FolderPath)
	if err != nil {
		return err
	}

	// Collect used packages from method signatures and dependencies
	usedPackages := make(map[string]bool)
	for _, service := range ctx.GeneratedCode.Services {
		// From method signatures
		for _, method := range service.Methods {
			collectPackagesFromType(method.ParamType, usedPackages)
			collectPackagesFromType(method.ReturnType, usedPackages)
		}
		// From dependencies
		for _, dep := range service.Dependencies {
			collectPackagesFromType(dep.InnerType, usedPackages)
		}
	}

	// Filter imports to only used packages
	allImports := make(map[string]string) // path -> alias
	for _, service := range ctx.GeneratedCode.Services {
		for alias, importPath := range service.Imports {
			// Skip if import path is "github.com/primadi/lokstra/core/service" (already added)
			if importPath == "github.com/primadi/lokstra/core/service" {
				continue
			}
			// Only include if package is actually used
			if usedPackages[alias] {
				// Use path as key to deduplicate, prefer shorter alias
				if existing, exists := allImports[importPath]; !exists || len(alias) < len(existing) {
					allImports[importPath] = alias
				}
			}
		}
	}

	// Generate code
	var buf bytes.Buffer
	if err := genTemplate.Execute(&buf, map[string]interface{}{
		"Package":           pkgName,
		"Services":          ctx.GeneratedCode.Services,
		"PreservedSections": ctx.GeneratedCode.PreservedSections,
		"AllImports":        allImports,
	}); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// collectPackagesFromType extracts package prefixes from type string
// e.g., "*domain.User" -> "domain", "[]userDomain.User" -> "userDomain"
func collectPackagesFromType(typeStr string, packages map[string]bool) {
	if typeStr == "" {
		return
	}

	// Remove pointer and array prefixes
	typeStr = strings.TrimLeft(typeStr, "*[]")

	// Extract package prefix (everything before last dot)
	if idx := strings.LastIndex(typeStr, "."); idx != -1 {
		pkg := typeStr[:idx]
		packages[pkg] = true
	}

	// Handle generics: Type[Param]
	if start := strings.Index(typeStr, "["); start != -1 {
		if end := strings.LastIndex(typeStr, "]"); end > start {
			innerTypes := typeStr[start+1 : end]
			// Split by comma for multiple type params
			for _, inner := range strings.Split(innerTypes, ",") {
				collectPackagesFromType(strings.TrimSpace(inner), packages)
			}
		}
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
		if strings.HasSuffix(file.Name(), "_test.go") || file.Name() == generatedFileName {
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

// extractStringArray extracts string array from interface{}
func extractStringArray(val interface{}) []string {
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

// genTemplate is the template for zz_generated.lokstra.go
var genTemplate = template.Must(template.New("gen").Funcs(template.FuncMap{
	"hasReturnValue":    hasReturnValue,
	"extractReturnType": extractReturnType,
	"quote":             func(s string) string { return fmt.Sprintf("%q", s) },
	"join":              strings.Join,
	"trimPrefix":        strings.TrimPrefix,
	"trimSuffix":        strings.TrimSuffix,
}).Parse(`// AUTO-GENERATED CODE - DO NOT EDIT
// Generated by lokstra-annotation from annotations in this folder
// Annotations: @RouterService, @Inject, @Route

package {{.Package}}

import (
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/core/service"
{{- range $path, $alias := .AllImports }}
	{{$alias}} "{{$path}}"
{{- end }}
)

// Auto-register on package import
func init() {
{{- range $name, $service := .Services }}
	Register{{$service.StructName}}()
{{- end }}
}

{{range $name, $service := .Services}}
// ============================================================
// FILE: {{$service.SourceFile}}
// ============================================================

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

{{range $method, $route := $service.Routes}}
{{- $sig := index $service.Methods $method}}
{{- if $sig}}
// {{$method}} via HTTP
// Generated from: @Route {{quote $route}}
{{if $sig.HasData}}func (s *{{$service.RemoteTypeName}}) {{$method}}(p {{$sig.ParamType}}) ({{$sig.ReturnType}}, error) {
	return proxy.CallWithData[{{$sig.ReturnType}}](s.proxyService, {{quote $method}}, p)
}
{{else}}func (s *{{$service.RemoteTypeName}}) {{$method}}(p {{$sig.ParamType}}) error {
	return proxy.Call(s.proxyService, {{quote $method}}, p)
}
{{end}}
{{- end}}
{{end}}

func {{$service.StructName}}Factory(deps map[string]any, config map[string]any) any {
	return &{{$service.StructName}}{
{{- range $svcName, $dep := $service.Dependencies }}
		{{$dep.FieldName}}: service.Cast[{{$dep.InnerType}}](deps[{{quote $dep.ServiceName}}]),
{{- end }}
	}
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
//   - @Inject annotations
//   - @Route annotations on methods
func Register{{$service.StructName}}() {
	// Register service type with router configuration
	lokstra_registry.RegisterServiceType("{{$service.ServiceName}}-factory",
		{{$service.StructName}}Factory,
		{{$service.RemoteTypeName}}Factory,
		deploy.WithRouter(&deploy.ServiceTypeRouter{
			PathPrefix:  {{quote $service.Prefix}},
			Middlewares: []string{ {{range $i, $mw := $service.Middlewares}}{{if $i}}, {{end}}{{quote $mw}}{{end}} },
			CustomRoutes: map[string]string{
{{- range $method, $route := $service.Routes }}
				{{quote $method}}:  {{quote $route}},
{{- end }}
			},
		}),
	)

	// Register lazy service with auto-detected dependencies
	lokstra_registry.RegisterLazyService({{quote $service.ServiceName}},
		"{{$service.ServiceName}}-factory",
		map[string]any{
			"depends-on": []string{ {{range $svcName, $dep := $service.Dependencies}}{{if ne $svcName ""}}{{quote $dep.ServiceName}}, {{end}}{{end}} },
		})
}
{{end}}
{{range $filename, $code := .PreservedSections}}
{{$code}}
{{end}}
`))

func hasReturnValue(route string) bool {
	return !strings.Contains(route, "error")
}

func extractReturnType(route string) string {
	return "*domain.User"
}
