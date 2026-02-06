package canonicalheader

import (
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var errNonConstantKey = errors.New("expression is not a constant key")

type constantString struct {
	originalValue,
	nameOfConst string

	pos token.Pos
	end token.Pos
}

func newConstantKey(info *types.Info, ident *ast.Ident) (constantString, error) {
	c, ok := info.ObjectOf(ident).(*types.Const)
	if !ok {
		return constantString{}, fmt.Errorf("%w: %T", errNonConstantKey, c)
	}

	return constantString{
		nameOfConst:   c.Name(),
		originalValue: constant.StringVal(c.Val()),
		pos:           ident.Pos(),
		end:           ident.End(),
	}, nil
}

func (c constantString) diagnostic(canonicalHeader string) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: c.pos,
		End: c.end,
		Message: fmt.Sprintf(
			"use canonical header key %q instead of %q",
			canonicalHeader,
			c.originalValue,
		),
	}
}

func (c constantString) value() string {
	return c.originalValue
}
