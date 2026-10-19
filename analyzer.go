package canonicalheader

import (
	"flag"
	"fmt"
	"go/ast"
	"maps"
	"net/http"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	httpHeaderPkg  = "net/http"
	httpHeaderName = "Header"
)

//nolint:gochecknoglobals // for backward compatibility.
var Analyzer = New()

func New() *analysis.Analyzer {
	c := &canonicalHeader{}

	a := &analysis.Analyzer{
		Name:     "canonicalheader",
		Doc:      "canonicalheader checks whether net/http.Header uses canonical header",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      c.run,
	}

	a.Flags.Init("canonicalheader", flag.ExitOnError)
	a.Flags.Var(&c.exclusions, "exclusions", "comma-separated list of exclusion rules")
	a.Flags.BoolVar(&c.useDefaultExclusion, "useDefaultExclusion", true, "use default exclusion rules")

	return a
}

type canonicalHeader struct {
	useDefaultExclusion bool
	exclusions          stringSet
}

func (c *canonicalHeader) buildExclusions() map[string]string {
	var defEx map[string]string
	if c.useDefaultExclusion {
		defEx = initialism()
	}

	result := make(map[string]string, len(c.exclusions)+len(defEx))

	maps.Copy(result, defEx)

	for _, ex := range c.exclusions {
		result[http.CanonicalHeaderKey(ex)] = ex
	}

	return result
}

func (c *canonicalHeader) run(pass *analysis.Pass) (any, error) {
	spctor, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, fmt.Errorf("want %T, got %T", spctor, pass.ResultOf[inspect.Analyzer])
	}

	exclusions := c.buildExclusions()

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	spctor.Preorder(nodeFilter, func(n ast.Node) {
		callExp, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		expr, ok := headerKeyExpr(pass.TypesInfo, callExp)
		if !ok {
			return
		}

		key, ok := readHeaderKey(pass.TypesInfo, expr)
		if !ok {
			return
		}

		canonical := canonicalHeaderKey(key.value, exclusions)
		if key.value == canonical {
			return
		}

		pass.Report(key.diagnostic(canonical))
	})

	//nolint:nilnil // This is expected by the signature of the callback.
	return nil, nil
}

func canonicalHeaderKey(s string, m map[string]string) string {
	canonical := http.CanonicalHeaderKey(s)

	wellKnown, ok := m[canonical]
	if !ok {
		return canonical
	}

	return wellKnown
}
