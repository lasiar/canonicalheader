package canonicalheader

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"unicode/utf8"
)

type headerKey struct {
	expr    ast.Expr
	value   string
	literal *ast.BasicLit
}

func readHeaderKey(info *types.Info, expr ast.Expr) (headerKey, bool) {
	value := info.Types[expr].Value
	if value == nil || value.Kind() != constant.String {
		return headerKey{}, false
	}

	key := headerKey{
		expr:    expr,
		value:   constant.StringVal(value),
		literal: stringLiteral(info, expr),
	}
	if !utf8.ValidString(key.value) {
		return headerKey{}, false
	}

	return key, true
}

func stringLiteral(info *types.Info, expr ast.Expr) *ast.BasicLit {
	for current := expr; ; {
		switch node := ast.Unparen(current).(type) {
		case *ast.BasicLit:
			if node.Kind == token.STRING {
				return node
			}

			return nil

		case *ast.CallExpr:
			if len(node.Args) != 1 || !info.Types[node.Fun].IsType() {
				return nil
			}

			current = node.Args[0]

		default:
			return nil
		}
	}
}
