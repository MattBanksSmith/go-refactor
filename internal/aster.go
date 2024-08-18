package internal

import (
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"os"
	"path/filepath"
)

// eventually pass config into the application to drive refactor.
// Not yet implemented - need abstraction to form and work out the contract
type config struct {
}

type configItem struct {
	nodeType  string //type e.g. funcDecl
	operation string //replace, add or replace, rename

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
		panic(err)
	}

	file, err := parser.ParseFile(fileSet, filePath, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing file %s: %v\n", filePath, err)
		return
	}

	// Traverse the AST and apply function on each
	astutil.Apply(file, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.FuncDecl:
			handleFuncDecl(astContext[*ast.FuncDecl]{
				node:     n,
				fileSet:  fileSet,
				filePath: filePath,
				file:     file,
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
