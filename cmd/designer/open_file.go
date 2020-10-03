package main

import (
	"errors"
	"github.com/gonutz/wui"
	"go/ast"
	"go/parser"
	"go/token"
)

// TODO openFile should return a slice of windows instead.
func openFile(path string) ([]*wui.Window, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	wuiImport, err := findWuiPackageImport(f)
	if err != nil {
		return nil, err
	}
	if wuiImport == "." {
		return nil, errors.New(
			"wui is imported as . which is currently not supported",
		)
	}
	creations := findWuiWindowCreationPoints(f, wuiImport)
	if len(creations) == 0 {
		return nil, errors.New("the file does not contain a wui.NewWindow() statement")
	}
	windows := make([]*wui.Window, len(creations))
	for i := range windows {
		windows[i], err = createWindow(creations[i])
		if err != nil {
			return nil, err
		}
	}
	return nil, errors.New("TODO")
}

func findWuiPackageImport(f *ast.File) (importName string, err error) {
	importName = "wui"
	found := false
	for _, imp := range f.Imports {
		if imp.Path.Value == `"github.com/gonutz/wui"` {
			if found {
				return "", errors.New("wui is imported multiple times")
			}
			found = true
			if imp.Name != nil {
				importName = imp.Name.Name
			}
		}
	}
	if !found {
		return "", errors.New("wui import was not found")
	}
	return importName, nil
}

type wuiWindowCreation struct {
	varName  string
	creation *ast.AssignStmt
	block    *ast.BlockStmt
}

func findWuiWindowCreationPoints(f *ast.File, wuiImport string) []wuiWindowCreation {
	var windows []wuiWindowCreation
	var lastBlock *ast.BlockStmt
	ast.Inspect(f, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			lastBlock = block
		}
		if name, ok := isWuiWindowCreation(wuiImport, f, n); ok {
			windows = append(windows, wuiWindowCreation{
				varName:  name,
				creation: n.(*ast.AssignStmt),
				block:    lastBlock,
			})
		}
		return true
	})
	return windows
}

func isWuiWindowCreation(wuiName string, f *ast.File, n ast.Node) (varName string, allOK bool) {
	if false {
	} else if assign, ok := n.(*ast.AssignStmt); !ok {
	} else if !(len(assign.Lhs) == 1) {
	} else if variable, ok := assign.Lhs[0].(*ast.Ident); !ok {
	} else if !(variable.Name != "_") {
	} else if !(len(assign.Rhs) == 1) {
	} else if call, ok := assign.Rhs[0].(*ast.CallExpr); !ok {
	} else if !(len(call.Args) == 0) {
	} else if sel, ok := call.Fun.(*ast.SelectorExpr); !ok {
	} else if pkg, ok := sel.X.(*ast.Ident); !ok {
	} else if !(pkg.Name == wuiName) {
	} else if !(sel.Sel.Name == "NewWindow") {
	} else if !containsIdent(f.Unresolved, pkg) {
	} else {
		return variable.Name, true
	}
	return
}

func containsIdent(ids []*ast.Ident, id *ast.Ident) bool {
	for i := range ids {
		if id == ids[i] {
			return true
		}
	}
	return false
}

func createWindow(c wuiWindowCreation) (*wui.Window, error) {
	for _, stmt := range c.block.List {
		// TODO
		_ = stmt
	}
	return nil, errors.New("TODO")
}
