package utils

import (
	"encoding/csv"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// CodeEntity represents a function or method in the codebase
type CodeEntity struct {
	Name       string
	Type       string // "function" or "method"
	FilePath   string
	Complexity int
}

// calculateComplexity calculates the cyclomatic complexity of an AST node
func calculateComplexity(n ast.Node) int {
	// Start with complexity of 1 (the basic path through the function)
	complexity := 1

	// Define a visitor that increments complexity for each control flow statement
	visitor := &complexityVisitor{complexity: &complexity}
	ast.Walk(visitor, n)

	return complexity
}

// complexityVisitor is an ast.Visitor that counts control flow statements
type complexityVisitor struct {
	complexity *int
}

func (v *complexityVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	// Increment complexity for control flow statements
	switch n := node.(type) {
	case *ast.IfStmt:
		*v.complexity++
	case *ast.ForStmt:
		*v.complexity++
	case *ast.RangeStmt:
		*v.complexity++
	case *ast.CaseClause:
		if n.List != nil { // Skip 'default' case
			*v.complexity++
		}
	case *ast.CommClause:
		if n.Comm != nil { // Skip 'default' case
			*v.complexity++
		}
	case *ast.BinaryExpr:
		// Count && and || operators as they create additional paths
		if n.Op == token.LAND || n.Op == token.LOR {
			*v.complexity++
		}
	}

	return v
}

func RunComplexityAnalyzer(rootDir string, outputPath string) {
	// Validate the root directory
	if rootDir == "" {
		rootDir = ".."
	}

	// Validate the output path
	if outputPath == "" {
		outputPath = "complexity_analysis.csv"
	}

	// Create a slice to store code entities
	var entities []CodeEntity

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
			// Skip hidden directories and vendor, but not the root directory or parent directory
			if path != "." && path != ".." && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build") {
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

		// Process declarations in the file
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				// Calculate complexity for the function body
				complexity := calculateComplexity(d.Body)

				// Determine if it's a method or a function
				entityType := "function"
				entityName := d.Name.Name

				if d.Recv != nil && len(d.Recv.List) > 0 {
					// It's a method, find the receiver type
					entityType = "method"

					// Get the receiver type
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

					// Format the method name as ReceiverType.MethodName
					if receiverType != "" {
						entityName = receiverType + "." + entityName
					}
				}

				// Add the entity to our list
				entities = append(entities, CodeEntity{
					Name:       entityName,
					Type:       entityType,
					FilePath:   path,
					Complexity: complexity,
				})
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		return
	}

	// Create the CSV file
	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CSV file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header
	header := []string{"Name", "Type", "FilePath", "Complexity"}
	if err := writer.Write(header); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV header: %v\n", err)
		return
	}

	// Write the data
	for _, entity := range entities {
		record := []string{
			entity.Name,
			entity.Type,
			entity.FilePath,
			fmt.Sprintf("%d", entity.Complexity),
		}
		if err := writer.Write(record); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing CSV record: %v\n", err)
			return
		}
	}

	fmt.Printf("Complexity analysis complete. Output written to %s\n", outputPath)
}
