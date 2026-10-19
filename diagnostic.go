package canonicalheader

import (
	"fmt"
	"go/ast"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

func (key headerKey) diagnostic(canonical string) analysis.Diagnostic {
	diagnostic := analysis.Diagnostic{
		Pos:     key.expr.Pos(),
		End:     key.expr.End(),
		Message: fmt.Sprintf("use %q instead of %q", canonical, key.value),
	}
	if key.literal != nil {
		diagnostic.Pos = key.literal.Pos()
		diagnostic.End = key.literal.End()
		diagnostic.SuggestedFixes = []analysis.SuggestedFix{
			literalFix(key.literal, key.value, canonical),
		}
	}

	return diagnostic
}

func literalFix(literal *ast.BasicLit, original, canonical string) analysis.SuggestedFix {
	text := strconv.Quote(canonical)
	if literal.Value[0] == '`' && strconv.CanBackquote(canonical) {
		text = "`" + canonical + "`"
	}

	return analysis.SuggestedFix{
		Message: fmt.Sprintf("should be replaced %q with %q", original, canonical),
		TextEdits: []analysis.TextEdit{
			{
				Pos:     literal.Pos(),
				End:     literal.End(),
				NewText: []byte(text),
			},
		},
	}
}
