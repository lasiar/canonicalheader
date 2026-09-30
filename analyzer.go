package canonicalheader

import (
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"net/http"

	"github.com/go-toolsmith/astcast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

const (
	pkgPath = "net/http"
	name    = "Header"
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

type argumenter interface {
	diagnostic(canonicalHeader string) analysis.Diagnostic
	value() string
}

func (c *canonicalHeader) buildExclusions() map[string]string {
	var defEx map[string]string
	if c.useDefaultExclusion {
		defEx = initialism()
	}

	result := make(map[string]string, len(c.exclusions)+len(defEx))

	for canonical, noCanonical := range defEx {
		result[canonical] = noCanonical
	}

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
		(*ast.IndexExpr)(nil),
		(*ast.CompositeLit)(nil),
	}
	var outerErr error

	spctor.Preorder(nodeFilter, func(n ast.Node) {
		if outerErr != nil {
			return
		}

		switch t := n.(type) {
		// h["Test-Header"].
		case *ast.IndexExpr:
			if isHeader(pass.TypesInfo.TypeOf(t.X)) {
				checkMapKey(pass, t.Index)
			}

			return

		// http.Header{"Test-Header": nil}.
		case *ast.CompositeLit:
			if !isHeader(pass.TypesInfo.TypeOf(t)) {
				return
			}

			for _, elt := range t.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					checkMapKey(pass, kv.Key)
				}
			}

			return
		}

		callExp, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		var (
			// gotType type receiver.
			gotType       types.Type
			gotMethodName string
		)

		switch t := typeutil.Callee(pass.TypesInfo, callExp).(type) {
		// Direct call method.
		case *types.Func:
			fn := t
			// Find net/http.Header{} by function call.
			signature, ok := fn.Type().(*types.Signature)
			if !ok {
				return
			}

			recv := signature.Recv()

			// It's a func, not a method.
			if recv == nil {
				return
			}
			gotType = recv.Type()
			gotMethodName = astcast.ToSelectorExpr(callExp.Fun).Sel.Name

		// h := http.Header{}
		// f := h.Get
		// v("Test-Value").
		case *types.Var:
			ident, ok := callExp.Fun.(*ast.Ident)
			if !ok {
				return
			}

			if ident.Obj == nil {
				return
			}

			// f := h.Get.
			assign, ok := ident.Obj.Decl.(*ast.AssignStmt)
			if !ok {
				return
			}

			// For case `i, v := 0, h.Get`.
			// indexAssign--^.
			indexAssign := -1
			for i, lh := range assign.Lhs {
				// Find by name of variable.
				if astcast.ToIdent(lh).Name == ident.Name {
					indexAssign = i
				}
			}

			// Not found.
			if indexAssign == -1 {
				return
			}

			if len(assign.Rhs) <= indexAssign {
				return
			}

			sel, ok := assign.Rhs[indexAssign].(*ast.SelectorExpr)
			if !ok {
				return
			}

			gotMethodName = sel.Sel.Name
			ident, ok = sel.X.(*ast.Ident)
			if !ok {
				return
			}

			obj := pass.TypesInfo.ObjectOf(ident)
			gotType = obj.Type()

		// delete(h, "Test-Header").
		case *types.Builtin:
			if t.Name() == "delete" && len(callExp.Args) == 2 && isHeader(pass.TypesInfo.TypeOf(callExp.Args[0])) {
				checkMapKey(pass, callExp.Args[1])
			}

			return

		default:
			return
		}

		// It is not net/http.Header{}.
		if !isHeader(gotType) {
			return
		}

		// Search for known methods where the key is the first arg.
		if !isValidMethod(gotMethodName) {
			return
		}

		// Should be more than one. Because get the value by index it
		// will not be superfluous.
		if len(callExp.Args) == 0 {
			return
		}

		arg, ok := newArgument(pass.TypesInfo, callExp.Args[0])
		if !ok {
			return
		}

		argValue := arg.value()

		headerKeyCanonical := canonicalHeaderKey(argValue, exclusions)
		if argValue == headerKeyCanonical {
			return
		}

		pass.Report(arg.diagnostic(headerKeyCanonical))
	})

	return nil, outerErr
}

// checkMapKey reports keys that bypass canonicalization through direct map
// access: exclusions do not apply, since any other spelling is a silent miss.
func checkMapKey(pass *analysis.Pass, key ast.Expr) {
	arg, ok := newArgument(pass.TypesInfo, key)
	if !ok {
		return
	}

	argValue := arg.value()

	headerKeyCanonical := http.CanonicalHeaderKey(argValue)
	if argValue == headerKeyCanonical {
		return
	}

	pass.Report(arg.diagnostic(headerKeyCanonical))
}

func newArgument(info *types.Info, expr ast.Expr) (argumenter, bool) {
	// Check for type casting from myString to string.
	// it could be: Get(string(string(string(myString)))).
	// need this node------------------------^^^^^^^^.
	for {
		// If it is not *ast.CallExpr, this is a value.
		c, ok := expr.(*ast.CallExpr)
		if !ok {
			break
		}

		// Some function is called, skip this case.
		if len(c.Args) == 0 {
			return nil, false
		}

		f, ok := c.Fun.(*ast.Ident)
		if !ok {
			break
		}

		obj := info.ObjectOf(f)
		// nil may be by code, but not by logic.
		// TypeInfo should contain of type.
		if obj == nil {
			break
		}

		// This is function.
		// Skip this method call.
		_, ok = obj.Type().(*types.Signature)
		if ok {
			return nil, false
		}

		expr = c.Args[0]
	}

	switch t := expr.(type) {
	case *ast.BasicLit:
		lString, err := newLiteralString(t)
		if err != nil {
			return nil, false
		}

		return lString, true

	case *ast.Ident:
		constString, err := newConstantKey(info, t)
		if err != nil {
			return nil, false
		}

		return constString, true

	default:
		return nil, false
	}
}

func canonicalHeaderKey(s string, m map[string]string) string {
	canonical := http.CanonicalHeaderKey(s)

	wellKnown, ok := m[canonical]
	if !ok {
		return canonical
	}

	return wellKnown
}

// isHeader reports whether t is net/http.Header, possibly through an alias.
func isHeader(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()

	return obj.Pkg() != nil && obj.Pkg().Path() == pkgPath && obj.Name() == name
}

func isValidMethod(name string) bool {
	switch name {
	case "Get", "Set", "Add", "Del", "Values":
		return true
	default:
		return false
	}
}
