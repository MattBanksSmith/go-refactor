package internal

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
)

type astContext[T any] struct {
	node     T
	fileSet  *token.FileSet
	filePath string
	file     *ast.File
	printer  *printer.Config
}

func checkContextName(data astContext[*ast.FuncDecl]) {
	for _, param := range data.node.Type.Params.List {
		for _, ident := range param.Names {
			if ident.Name == "context" {
				fmt.Printf("File %s: Function %s has argument named 'context', should be 'ctx'\n", data.filePath, data.node.Name.Name)
			}
		}
	}
}

// injectContext rewrites a function signature to have context as the first argument
func injectContext(data astContext[*ast.FuncDecl], functionList map[string]map[string]struct{}) error {
	for _, param := range data.node.Type.Params.List {
		switch t := param.Type.(type) {
		case *ast.SelectorExpr:
			x, ok := t.X.(*ast.Ident)
			if ok && x.Name == "context" && t.Sel.Name == "Context" {
				return nil
			}
		}
	}

	newParam := createCtxParam(data)

	// Force context as the first argument
	data.node.Type.Params.List = append([]*ast.Field{newParam}, data.node.Type.Params.List...)

	if !importExists(data.file, "context") {
		astutil.AddImport(data.fileSet, data.file, "context")
	}

	file, err := os.Create(data.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = data.printer.Fprint(file, data.fileSet, data.file); err != nil {
		return err
	}

	return nil
}

func createCtxParam(data astContext[*ast.FuncDecl]) *ast.Field {
	// Todo almost certainly a bug with the positions here, e.g. multi line function declarations
	// pos is necessary to set to avoid trailing commas. When unset the line number is 0 causing nodes.go to
	// incorrectly identify if its a multiline function or not
	// trailing comma bug https://github.com/golang/go/issues/23771
	nameIdent := ast.NewIdent("ctx")
	nameIdent.NamePos = data.node.Pos()

	argIdent := ast.NewIdent("context")
	argIdent.NamePos = nameIdent.End()

	selIdent := ast.NewIdent("Context")
	selIdent.NamePos = argIdent.End()

	newParam := &ast.Field{
		Names: []*ast.Ident{nameIdent},
		Type: &ast.SelectorExpr{
			X:   argIdent,
			Sel: selIdent,
		},
	}
	return newParam
}

func importExists(f *ast.File, name string) bool {
	for _, imp := range f.Imports {
		if imp.Path.Value == `"`+name+`"` {
			return true
		}
	}
	return false
}
