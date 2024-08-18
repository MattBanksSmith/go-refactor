package main

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"os"
)

type stuff[T any] struct {
	node     T
	fileSet  *token.FileSet
	filePath string
	file     *ast.File
}

func checkContextName(data stuff[*ast.FuncDecl]) {
	for _, param := range data.node.Type.Params.List {
		for _, ident := range param.Names {
			if ident.Name == "context" {
				fmt.Printf("File %s: Function %s has argument named 'context', should be 'ctx'\n", data.filePath, data.node.Name.Name)
			}
		}
	}
}

// injectContext rewrites a function signature to have context as the first argument
func injectContext(data stuff[*ast.FuncDecl], functionList map[string]map[string]struct{}) {
	for _, param := range data.node.Type.Params.List {
		if ident, ok := param.Type.(*ast.Ident); ok && ident.Name == "context.Context" {
			//context exits, do nothing
			return
		}
	}

	file, err := os.Create(data.filePath)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	if err = printer.Fprint(file, data.fileSet, data.node); err != nil {
		log.Println(err)
	}
}
