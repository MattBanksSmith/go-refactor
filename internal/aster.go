package internal

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
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

func getFuncList() (map[string]struct{}, error) {
	f, err := os.Open("data.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	result := make(map[string]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		result[line] = struct{}{}
	}
	return result, nil
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

	funcList, err := getFuncList()
	if err != nil {
		log.Printf("Error getting function list: %v\n", err)
		return
	}

	// Traverse the AST and apply function on each
	astutil.Apply(file, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.FuncDecl:
			name := getFullName(n)
			pkg, err := getImportPath(filePath)
			if err != nil {
				log.Printf("Error getting import path for %s: %v\n", filePath, err)
			}
			fmt.Println(pkg + "." + name)
			if _, ok := funcList[pkg+"."+name]; !ok {
				return true
			}

			handleFuncDecl(astContext[*ast.FuncType]{
				node:     n.Type,
				fileSet:  fileSet,
				filePath: filePath,
				file:     file,
				printer:  &printer.Config{},
			})
		case *ast.FuncLit:
			//Todo how on earth to handle anonymous funcs?
			/*handleFuncDecl(astContext[*ast.FuncType]{
				node:     n.Type,
				fileSet:  fileSet,
				filePath: filePath,
				file:     file,
				printer:  &printer.Config{},
			})*/
		}
		return true
	}, nil)
}

func getImportPath(filePath string) (string, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, fmt.Sprintf("file=%s", filePath))
	if err != nil {
		return "", err
	}

	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package found for file: %s", filePath)
	}

	return pkgs[0].PkgPath, nil
}

func getFullName(fn *ast.FuncDecl) string {
	var fullName string
	if fn.Recv != nil {
		// receiver method
		recvType := "unknown"
		if len(fn.Recv.List) > 0 {
			switch expr := fn.Recv.List[0].Type.(type) {
			case *ast.Ident:
				recvType = expr.Name
			case *ast.StarExpr:
				if ident, ok := expr.X.(*ast.Ident); ok {
					recvType = ident.Name
				}
			}
		}
		fullName = fmt.Sprintf("%s.%s", recvType, fn.Name.Name)
	} else {
		// Function
		fullName = fn.Name.Name
	}
	return fullName
}

func handleFuncDecl(data astContext[*ast.FuncType]) {
	checkContextName(data)
	err := injectContext(data, map[string]map[string]struct{}{})
	if err != nil {
		log.Println(err)
	}
}
