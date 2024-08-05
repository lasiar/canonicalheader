package _const

import "net/http"

const (
	// noUsage for testing ast.
	noUsage     = ""
	noCanonical = `TT`
	canonical   = "Tt"
)

const copiedFromNoCanonical = noCanonical

type myString string

const underlyingString myString = "TT"

func _() {
	var mstr myString = "Tt"
	http.Header{}.Get(string(mstr))

	http.Header{}.Get(string(underlyingString)) // want `use "Tt" instead of "TT"`
	http.Header{}.Get(string(underlyingString)) // want `use "Tt" instead of "TT"`
	http.Header{}.Get(noCanonical)              // want `use "Tt" instead of "TT"`
	http.Header{}.Get(copiedFromNoCanonical)    // want `use "Tt" instead of "TT"`
	http.Header{}.Get(canonical)
}
