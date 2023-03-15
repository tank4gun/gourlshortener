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
	"honnef.co/go/tools/staticcheck"
)

// NoExitAnalyzer - analyzer for os.Exit() calls in main package
var NoExitAnalyzer = &analysis.Analyzer{
	Name: "noexitcheck",             // Name - analyzer name
	Doc:  "check for os.Exit usage", // Doc - docstring for analyzer
	Run:  run,                       // Run - method for analyzer run
}

// run - analyzer start function
func run(pass *analysis.Pass) (interface{}, error) {
	expr := func(x *ast.ExprStmt) {
		if call, ok := x.X.(*ast.CallExpr); ok {
			if s, ok := call.Fun.(*ast.SelectorExpr); ok {
				if s.Sel.Name == "Exit" {
					pass.Reportf(x.Pos(), "expression is os.Exit()")
				}
			}
		}
	}
	for _, file := range pass.Files {
		if file.Name.Name != "main.go" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.ExprStmt:
				expr(x)
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
		NoExitAnalyzer,
	}
	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Name, "SA") || v.Name == "S1010" || v.Name == "S1019" {
			mychecks = append(mychecks, v)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
