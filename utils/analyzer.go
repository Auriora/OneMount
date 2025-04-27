package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ModuleOutput represents the output format required for the module analysis
type ModuleOutput struct {
	ModuleName   string   `json:"moduleName"`
	Description  string   `json:"description"`
	FileCount    int      `json:"fileCount"`
	Dependencies []string `json:"dependencies"`
}

// ModuleInfo represents information about a Go module
type ModuleInfo struct {
	Name         string                 `json:"name"`
	Path         string                 `json:"path"`
	Classes      []ClassInfo            `json:"classes,omitempty"`
	Functions    []FunctionInfo         `json:"functions,omitempty"`
	Dependencies map[string]interface{} `json:"dependencies,omitempty"`
	FileCount    int                    `json:"-"`
}

// ClassInfo represents information about a Go struct (class)
type ClassInfo struct {
	Name       string       `json:"name"`
	Fields     []FieldInfo  `json:"fields,omitempty"`
	Methods    []MethodInfo `json:"methods,omitempty"`
	Implements []string     `json:"implements,omitempty"`
	Embedded   []string     `json:"embedded,omitempty"`
	Path       string       `json:"path"`
}

// FieldInfo represents a field in a struct
type FieldInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// MethodInfo represents a method of a struct
type MethodInfo struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters,omitempty"`
	Returns    []string `json:"returns,omitempty"`
}

// FunctionInfo represents a standalone function
type FunctionInfo struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters,omitempty"`
	Returns    []string `json:"returns,omitempty"`
}

// ImportInfo tracks imports in a file
type ImportInfo struct {
	Path  string
	Alias string
}

// generateModuleDescription creates a meaningful one-sentence description for a module
func generateModuleDescription(module *ModuleInfo) string {
	// Map of known module paths to their descriptions
	knownModules := map[string]string{
		"cmd/common":             "Common utilities and configuration shared across command-line applications",
		"cmd/onedriver":          "Main filesystem application for Microsoft OneDrive integration",
		"cmd/onedriver-launcher": "GUI launcher application for the onedriver filesystem",
		"fs":                     "Core filesystem implementation for OneDrive integration",
		"fs/graph":               "Microsoft Graph API client for OneDrive communication",
		"fs/offline":             "Offline mode functionality for the filesystem",
		"ui":                     "Graphical user interface components for onedriver",
		"ui/systemd":             "Systemd integration for the UI components",
		"testutil":               "Testing utilities for onedriver components",
	}

	// Check if we have a predefined description
	if desc, ok := knownModules[module.Path]; ok {
		return desc
	}

	// Generate a description based on the module path
	parts := strings.Split(module.Path, "/")
	lastPart := parts[len(parts)-1]

	// Replace underscores and hyphens with spaces and capitalize
	lastPart = strings.ReplaceAll(lastPart, "_", " ")
	lastPart = strings.ReplaceAll(lastPart, "-", " ")

	// Generate a generic description
	return fmt.Sprintf("Module providing %s functionality", lastPart)
}

func main() {
	// Define the root directory to analyze
	rootDir := ".."

	// Create a map to store module information
	modules := make(map[string]*ModuleInfo)

	// Create a file set for parsing
	fset := token.NewFileSet()

	// Walk through the directory structure
	fmt.Fprintln(os.Stderr, "Starting directory walk...")
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing path %s: %v\n", path, err)
			return err
		}

		// Skip non-Go files and directories
		if info.IsDir() {
			fmt.Fprintf(os.Stderr, "Found directory: %s\n", path)
			// Skip hidden directories and vendor, but not the root directory or parent directory
			if path != "." && path != ".." && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build") {
				fmt.Fprintf(os.Stderr, "Skipping directory: %s\n", path)
				return filepath.SkipDir
			}
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		fmt.Fprintf(os.Stderr, "Processing Go file: %s\n", path)

		// Parse the Go file
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", path, err)
			return nil
		}

		// Get the package name
		packageName := file.Name.Name
		packagePath := filepath.Dir(path)

		// Create or get the module info
		module, exists := modules[packagePath]
		if !exists {
			module = &ModuleInfo{
				Name:         packageName,
				Path:         packagePath,
				Classes:      []ClassInfo{},
				Functions:    []FunctionInfo{},
				Dependencies: make(map[string]interface{}),
				FileCount:    0,
			}
			modules[packagePath] = module
		}

		// Increment file count for this module
		module.FileCount++

		// Track imports
		imports := make(map[string]ImportInfo)
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			var alias string
			if imp.Name != nil {
				alias = imp.Name.Name
			} else {
				parts := strings.Split(importPath, "/")
				alias = parts[len(parts)-1]
			}
			imports[alias] = ImportInfo{Path: importPath, Alias: alias}
			module.Dependencies[importPath] = true
		}

		// Process declarations in the file
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok == token.TYPE {
					for _, spec := range d.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							// Check if it's a struct
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								// Found a struct (class)
								class := ClassInfo{
									Name:   typeSpec.Name.Name,
									Fields: []FieldInfo{},
									Path:   path,
								}

								// Process struct fields
								if structType.Fields != nil {
									for _, field := range structType.Fields.List {
										for _, name := range field.Names {
											fieldType := ""
											if expr, ok := field.Type.(*ast.Ident); ok {
												fieldType = expr.Name
											} else {
												fieldType = fmt.Sprintf("%T", field.Type)
											}
											class.Fields = append(class.Fields, FieldInfo{
												Name: name.Name,
												Type: fieldType,
											})
										}

										// Check for embedded types
										if len(field.Names) == 0 {
											if ident, ok := field.Type.(*ast.Ident); ok {
												class.Embedded = append(class.Embedded, ident.Name)
											} else if sel, ok := field.Type.(*ast.SelectorExpr); ok {
												if x, ok := sel.X.(*ast.Ident); ok {
													class.Embedded = append(class.Embedded, x.Name+"."+sel.Sel.Name)
												}
											}
										}
									}
								}

								module.Classes = append(module.Classes, class)
							}
						}
					}
				}
			case *ast.FuncDecl:
				// Check if it's a method
				if d.Recv != nil && len(d.Recv.List) > 0 {
					// It's a method, find the receiver type
					var receiverType string
					field := d.Recv.List[0]

					switch expr := field.Type.(type) {
					case *ast.StarExpr:
						if ident, ok := expr.X.(*ast.Ident); ok {
							receiverType = ident.Name
						}
					case *ast.Ident:
						receiverType = expr.Name
					}

					if receiverType != "" {
						// Find the class this method belongs to
						for i, class := range module.Classes {
							if class.Name == receiverType {
								method := MethodInfo{
									Name:       d.Name.Name,
									Parameters: []string{},
									Returns:    []string{},
								}

								// Process parameters
								if d.Type.Params != nil {
									for _, param := range d.Type.Params.List {
										paramType := fmt.Sprintf("%T", param.Type)
										method.Parameters = append(method.Parameters, paramType)
									}
								}

								// Process return values
								if d.Type.Results != nil {
									for _, result := range d.Type.Results.List {
										resultType := fmt.Sprintf("%T", result.Type)
										method.Returns = append(method.Returns, resultType)
									}
								}

								module.Classes[i].Methods = append(module.Classes[i].Methods, method)
								break
							}
						}
					}
				} else {
					// It's a standalone function
					function := FunctionInfo{
						Name:       d.Name.Name,
						Parameters: []string{},
						Returns:    []string{},
					}

					// Process parameters
					if d.Type.Params != nil {
						for _, param := range d.Type.Params.List {
							paramType := fmt.Sprintf("%T", param.Type)
							function.Parameters = append(function.Parameters, paramType)
						}
					}

					// Process return values
					if d.Type.Results != nil {
						for _, result := range d.Type.Results.List {
							resultType := fmt.Sprintf("%T", result.Type)
							function.Returns = append(function.Returns, resultType)
						}
					}

					module.Functions = append(module.Functions, function)
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		return
	}

	// Convert the map to the required output format
	var outputModules []ModuleOutput
	for _, module := range modules {
		// Only include modules that define at least one function or class
		if len(module.Functions) > 0 || len(module.Classes) > 0 {
			// Convert dependencies map to a slice
			deps := make([]string, 0, len(module.Dependencies))
			for dep := range module.Dependencies {
				deps = append(deps, dep)
			}

			// Generate a meaningful description for the module
			description := generateModuleDescription(module)

			// Create the output module with a more specific module name
			moduleName := module.Name
			if module.Path != "." {
				// Use the path as part of the module name to avoid duplicates
				moduleName = module.Path
				// Clean up the module name by removing the "../" prefix
				moduleName = strings.TrimPrefix(moduleName, "../")
			}

			outputModule := ModuleOutput{
				ModuleName:   moduleName,
				Description:  description,
				FileCount:    module.FileCount,
				Dependencies: deps,
			}

			outputModules = append(outputModules, outputModule)
		}
	}

	// Output the result as JSON
	jsonData, err := json.MarshalIndent(outputModules, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling to JSON: %v\n", err)
		return
	}

	// Write the output to a file in the docs/ folder
	outputPath := "../docs/onedriver_modules.json"
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file %s: %v\n", outputPath, err)
		return
	}

	fmt.Printf("Module analysis complete. Output written to %s\n", outputPath)
}
