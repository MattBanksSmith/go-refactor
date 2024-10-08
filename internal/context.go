package internal

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/ast/astutil"
	"os"
)

func checkContextName(data astContext[*ast.FuncType]) {
	for _, param := range data.node.Params.List {
		for _, ident := range param.Names {
			if ident.Name == "context" {
				fmt.Printf("File %s: Function %s has argument named 'context', should be 'ctx'\n", data.filePath, data.node)
			}
		}
	}
}

// injectContext rewrites a function signature to have context as the first argument
func injectContext(data astContext[*ast.FuncType], functionList map[string]map[string]struct{}) error {

	for _, param := range data.node.Params.List {
		switch t := param.Type.(type) {
		case *ast.SelectorExpr:
			x, ok := t.X.(*ast.Ident)
			if ok && x.Name == "context" && t.Sel.Name == "Context" {
				if len(param.Names) > 1 {
					return nil
				}

				if param.Names[0].Name == "ctx" {
					return nil
				}

				param.Names[0].Name = "ctx"
				err := writeData(data)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}

	newParam := createCtxParam(data)

	// Force context as the first argument
	data.node.Params.List = append([]*ast.Field{newParam}, data.node.Params.List...)

	if !importExists(data.file, "context") {
		astutil.AddImport(data.fileSet, data.file, "context")
	}

	err := writeData(data)
	if err != nil {
		return err
	}

	return nil
}

func writeData(data astContext[*ast.FuncType]) error {
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

func createCtxParam(data astContext[*ast.FuncType]) *ast.Field {
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
