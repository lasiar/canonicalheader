package canonicalheader

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/go-toolsmith/astcast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

const (
	pkgPath = "net/http"
	name    = "Header"
)

var Analyzer = &analysis.Analyzer{
	Name: "canonicalHeader",
	Doc:  "canonicalHeader checks whether net/http.Header uses canonical header",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	var outerErr error
	inspect := func(node ast.Node) bool {
		if outerErr != nil {
			return false
		}

		callExp, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selExp, ok := callExp.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}

		object, ok := pass.TypesInfo.TypeOf(selExp.X).(*types.Named)
		if !ok {
			return false
		}

		if !isHTTPHeader(object) {
			return false
		}

		if !isValidMethod(astcast.ToSelectorExpr(callExp.Fun).Sel.Name) {
			return false
		}

		arg, ok := callExp.Args[0].(*ast.BasicLit)
		if !ok {
			return false
		}

		if arg.Kind != token.STRING {
			return true
		}

		if len(arg.Value) < 2 {
			return true
		}

		quote := arg.Value[0]
		headerKeyOriginal, err := strconv.Unquote(arg.Value)
		if err != nil {
			outerErr = err
			return true
		}

		headerKeyCanonical := http.CanonicalHeaderKey(headerKeyOriginal)
		if headerKeyOriginal == headerKeyCanonical {
			return true
		}

		newText := make([]byte, 0, len(headerKeyCanonical)+2)
		newText = append(newText, quote)
		newText = append(newText, unsafe.Slice(unsafe.StringData(headerKeyCanonical), len(headerKeyCanonical))...)
		newText = append(newText, quote)

		pass.Report(
			analysis.Diagnostic{
				Pos:     arg.Pos(),
				End:     arg.End(),
				Message: fmt.Sprintf("non-canonical header %q, instead use: %q", headerKeyOriginal, headerKeyCanonical),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: fmt.Sprintf("should replace %q with %q", headerKeyOriginal, headerKeyCanonical),
						TextEdits: []analysis.TextEdit{
							{
								Pos:     arg.Pos(),
								End:     arg.End(),
								NewText: newText,
							},
						},
					},
				},
			},
		)

		return true
	}

	for _, f := range pass.Files {
		if outerErr != nil {
			return nil, outerErr
		}

		if !astutil.UsesImport(f, pkgPath) {
			continue
		}

		ast.Inspect(f, inspect)
	}

	//nolint:nilnil // not need return.
	return nil, nil
}

func isHTTPHeader(named *types.Named) bool {
	return named.Obj() != nil &&
		named.Obj().Pkg() != nil &&
		named.Obj().Pkg().Path() == pkgPath &&
		named.Obj().Name() == name
}

func isValidMethod(name string) bool {
	switch name {
	case "Get", "Set", "Add", "Del", "Values":
		return true
	default:
		return false
	}
}
