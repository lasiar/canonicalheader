package canonicalheader

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/tools/go/analysis"
)

var (
	errLiteralNotString     = errors.New("literal is not a string")
	errLiteralInvalidLength = errors.New("literal has invalid value length")
	errLiteralInvalidQuote  = errors.New("literal has unsupported quote")
	errLiteralInvalidUTF8   = errors.New("literal contains invalid UTF-8")
)

type literalString struct {
	originalValue string
	quote         byte
	pos, end      token.Pos
}

func newLiteralString(basicList *ast.BasicLit) (literalString, error) {
	if basicList.Kind != token.STRING {
		return literalString{}, fmt.Errorf("%w: %#v", errLiteralNotString, basicList)
	}

	if len(basicList.Value) < 2 {
		return literalString{}, fmt.Errorf("%w: %#v, value %q", errLiteralInvalidLength, basicList, basicList.Value)
	}

	quote := basicList.Value[0]
	switch quote {
	case '`', '"':
	default:
		return literalString{}, fmt.Errorf("%w: %q", errLiteralInvalidQuote, quote)
	}

	originalValue, err := strconv.Unquote(basicList.Value)
	if err != nil {
		return literalString{}, fmt.Errorf("unquote %q: %w", basicList.Value, err)
	}

	if !utf8.ValidString(originalValue) {
		return literalString{}, fmt.Errorf("%w: %#v", errLiteralInvalidUTF8, basicList.Value)
	}

	return literalString{
		originalValue: originalValue,
		quote:         quote,
		pos:           basicList.Pos(),
		end:           basicList.End(),
	}, nil
}

func (l literalString) diagnostic(canonicalHeader string) analysis.Diagnostic {
	newText := make([]byte, 0, len(canonicalHeader)+2)
	newText = append(newText, l.quote)
	newText = append(newText, unsafe.Slice(unsafe.StringData(canonicalHeader), len(canonicalHeader))...)
	newText = append(newText, l.quote)

	return analysis.Diagnostic{
		Pos:     l.pos,
		End:     l.end,
		Message: fmt.Sprintf("use canonical header key %q instead of %q", canonicalHeader, l.originalValue),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: fmt.Sprintf("should be replaced %q with %q", l.originalValue, canonicalHeader),
				TextEdits: []analysis.TextEdit{
					{
						Pos:     l.pos,
						End:     l.end,
						NewText: newText,
					},
				},
			},
		},
	}
}

func (l literalString) value() string {
	return l.originalValue
}
