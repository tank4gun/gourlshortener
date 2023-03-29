// staticlint module checks that package passes all static checks presented here.
// Run `go build staticlint.go` in order to build staticlint binary and then `./staticlint .` or `./staticlint ./...` to run checks.
package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

// NoOsExitAnalyzer - analyzer for os.Exit() calls in main package
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noexitcheck",             // Name - analyzer name
	Doc:  "check for os.Exit usage", // Doc - docstring for analyzer
	Run:  run,                       // Run - method for analyzer run
}

// checkOsExitExpr - checks whether given ExprStmt is os.Exit() and reports it to pass
func checkOsExitExpr(pass *analysis.Pass, x *ast.ExprStmt) {
	if call, ok := x.X.(*ast.CallExpr); ok {
		if s, ok := call.Fun.(*ast.SelectorExpr); ok {
			if s.Sel.Name == "Exit" {
				pass.Reportf(x.Pos(), "expression is os.Exit()")
			}
		}
	}
}

// run - analyzer start function
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "main" {
					return true
				}
				return false
			case *ast.ExprStmt:
				checkOsExitExpr(pass, x)
			}
			return true
		})
	}
	return nil, nil
}

func main() {
	mychecks := []*analysis.Analyzer{
		bools.Analyzer,
		shadow.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		unreachable.Analyzer,
		NoOsExitAnalyzer,
	}
	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Name, "SA") || v.Name == "S1010" || v.Name == "S1019" {
			mychecks = append(mychecks, v)
		}
	}
	for _, v := range simple.Analyzers {
		if v.Name == "S1020" || v.Name == "S1034" {
			mychecks = append(mychecks, v)
		}
	}
	multichecker.Main(
		mychecks...,
	)
}
