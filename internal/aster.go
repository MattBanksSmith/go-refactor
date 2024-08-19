package internal

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"os"
	"path/filepath"
)

type astContext[T any] struct {
	node     T
	fileSet  *token.FileSet
	filePath string
	file     *ast.File
	printer  *printer.Config
}

func Do(dir string) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".go" {
			checkFile(path)
		}
		return nil
	})
	return err
}

func checkFile(filePath string) {
	// Parse the Go file
	fileSet := token.NewFileSet()
	src, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	file, err := parser.ParseFile(fileSet, filePath, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing file %s: %v\n", filePath, err)
		return
	}

	pkg, err := build.ImportDir(filepath.Dir(filePath), build.IgnoreVendor)
	if err != nil {
		log.Printf("Error importing directory %s: %v\n", filepath.Dir(filePath), err)
		return
	}

	//todo does this only work with gopath?
	fmt.Println(pkg.ImportPath)

	// Traverse the AST and apply function on each
	astutil.Apply(file, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.FuncDecl:
			handleFuncDecl(astContext[*ast.FuncDecl]{
				node:     n,
				fileSet:  fileSet,
				filePath: filePath,
				file:     file,
				printer:  &printer.Config{},
			})
		}
		return true
	}, nil)
}

func handleFuncDecl(data astContext[*ast.FuncDecl]) {
	checkContextName(data)
	err := injectContext(data, map[string]map[string]struct{}{})
	if err != nil {
		log.Println(err)
	}
}
