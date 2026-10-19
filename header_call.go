package canonicalheader

import (
	"go/ast"
	"go/types"

	"github.com/go-toolsmith/astcast"
	"golang.org/x/tools/go/types/typeutil"
)

func headerKeyExpr(info *types.Info, call *ast.CallExpr) (ast.Expr, bool) {
	if len(call.Args) == 0 {
		return nil, false
	}

	var receiver types.Type
	var method string

	switch callee := typeutil.Callee(info, call).(type) {
	case *types.Func:
		signature, ok := callee.Type().(*types.Signature)
		if !ok || signature.Recv() == nil {
			return nil, false
		}

		selector := astcast.ToSelectorExpr(call.Fun)
		if selector.Sel == nil {
			return nil, false
		}

		receiver = signature.Recv().Type()
		method = selector.Sel.Name

	case *types.Var:
		selector, ok := assignedMethod(call.Fun)
		if !ok {
			return nil, false
		}

		ident, ok := selector.X.(*ast.Ident)
		if !ok {
			return nil, false
		}

		receiver = info.TypeOf(ident)
		method = selector.Sel.Name

	default:
		return nil, false
	}

	if !isHTTPHeader(receiver) || !isValidMethod(method) {
		return nil, false
	}

	return call.Args[0], true
}

// assignedMethod resolves method values declared by assignments such as f := h.Get.
func assignedMethod(expr ast.Expr) (*ast.SelectorExpr, bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok || ident.Obj == nil {
		return nil, false
	}

	assign, ok := ident.Obj.Decl.(*ast.AssignStmt)
	if !ok {
		return nil, false
	}

	for i, lhs := range assign.Lhs {
		if astcast.ToIdent(lhs).Name != ident.Name {
			continue
		}

		if i >= len(assign.Rhs) {
			return nil, false
		}

		selector, ok := assign.Rhs[i].(*ast.SelectorExpr)
		if !ok || selector.Sel == nil {
			return nil, false
		}

		return selector, true
	}

	return nil, false
}

func isValidMethod(name string) bool {
	switch name {
	case "Get", "Set", "Add", "Del", "Values":
		return true
	default:
		return false
	}
}

func isHTTPHeader(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil {
		return false
	}

	if obj.Name() != httpHeaderName {
		return false
	}

	pkg := obj.Pkg()
	if pkg == nil {
		return false
	}

	return pkg.Path() == httpHeaderPkg
}
