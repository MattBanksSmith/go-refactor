package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func main() {
	// Specify the directory to scan for Go files
	dir := "sample"

	// Recursively scan the directory for Go files
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".go" {
			checkFile(path)
		}
		return nil
	})
}

func checkFile(filePath string) {
	// Parse the Go file
	fset := token.NewFileSet()
	src, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	f, err := parser.ParseFile(fset, filePath, src, parser.AllErrors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", filePath, err)
		return
	}

	// Traverse the AST and check for function arguments named "context"
	ast.Inspect(f, func(node ast.Node) bool {
		fmt.Println(f)
		switch n := node.(type) {
		case *ast.FuncDecl:
			handleFuncDecl(stuff[*ast.FuncDecl]{
				node:     n,
				fileSet:  fset,
				filePath: filePath,
				file:     f,
			})
		}
		return true
	})
}

func handleFuncDecl(data stuff[*ast.FuncDecl]) {
	checkContextName(data)
	injectContext(data, map[string]map[string]struct{}{})
}
