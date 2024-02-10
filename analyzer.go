package funcliner

import (
	"errors"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var errPositionInvalid = errors.New("position is invalid")

var Analyzer = &analysis.Analyzer{
	Name:     "funcliner",
	Doc:      "funcliener checks params func on a same line or on a separate line",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

type positioner interface {
	Pos() token.Pos
}

type position struct {
	p *analysis.Pass
}

func (pos position) startLine(node positioner) (int, error) {
	p := pos.p.Fset.Position(node.Pos())
	if !p.IsValid() {
		return 0, errPositionInvalid
	}

	return p.Line, nil
}

func run(pass *analysis.Pass) (any, error) {
	var outerErr error

	pos := position{p: pass}

	visitor := func(node ast.Node) bool {
		if outerErr != nil {
			return false
		}

		function, ok := node.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// name declaration here, for ease debug.
		name := function.Name.Name

		fStartAt, err := pos.startLine(function)
		if err != nil {
			outerErr = err
			return false
		}

		fEndParam := pass.Fset.Position(function.Type.Params.Closing).Line

		// func declaration and all params on a same line.
		if fStartAt == fEndParam {
			return false
		}

		wantLines := 2 + function.Type.Params.NumFields()

		gotLines := make(map[int]struct{}, wantLines)
		gotLines[fStartAt] = struct{}{}
		gotLines[fEndParam] = struct{}{}

		for _, param := range function.Type.Params.List {
			for _, ident := range param.Names {
				line, err := pos.startLine(ident)
				if err != nil {
					outerErr = err
					return false
				}

				gotLines[line] = struct{}{}
			}
		}

		if len(gotLines) == 1 || len(gotLines) >= wantLines {
			return false
		}

		pass.Reportf(function.Pos(), "%s params on one line or each parameter on a separate line", name)
		return true
	}

	for _, f := range pass.Files {
		ast.Inspect(f, visitor)
	}

	return nil, outerErr
}
