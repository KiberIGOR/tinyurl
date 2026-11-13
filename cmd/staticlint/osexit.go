package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer запрещает прямой вызов os.Exit в функции main пакета main.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc: `forbids a direct call to os.Exit in the main function of package main

A direct os.Exit in main skips deferred cleanup and makes the entry point
harder to test. Use a returned error, log.Fatal, or os.Exit outside main instead.

Only selector calls to os.Exit (including renamed imports of package os)
inside func main of package main are reported. Generated files and
synthetic *.test packages created by go test are skipped.`,
	Run: runOsExit,
}

func runOsExit(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	// go/packages создаёт синтетический пакет pkg.test с func main + os.Exit.
	if strings.HasSuffix(pass.Pkg.Path(), ".test") {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name == nil || fn.Name.Name != "main" || fn.Body == nil {
				return true
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if isOsExitCall(pass, call) {
					pass.Reportf(call.Pos(), "direct call to os.Exit in main function of package main is forbidden")
				}
				return true
			})
			return false
		})
	}
	return nil, nil
}

func isOsExitCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Exit" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	obj := pass.TypesInfo.ObjectOf(ident)
	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return false
	}
	return pkgName.Imported().Path() == "os"
}
