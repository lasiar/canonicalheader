package assigned

import (
	"fmt"
	"net/http"
	"strings"
)

func _() {
	h := http.Header{}

	i, g := 0, h.Del
	fmt.Println(i)
	g("TT") // want `use canonical header key "Tt" instead of "TT"`

	f := h.Get
	f("TT") // want `use canonical header key "Tt" instead of "TT"`

	var v = h.Get
	v("TT") // want `use canonical header key "Tt" instead of "TT"`

	getter := h.Get
	getter = strings.ToUpper
	getter("TT")

	if byIfInit := h.Get; true {
		byIfInit("TT") // want `use canonical header key "Tt" instead of "TT"`
	}

	if byIfInit := h.Get; true {
		byIfInit = strings.ToUpper
		byIfInit("TT")
	}

	for byForInit := h.Get; ; {
		byForInit("TT") // want `use canonical header key "Tt" instead of "TT"`
		break
	}

	for byForInit := h.Get; ; {
		byForInit = strings.ToUpper
		byForInit("TT")
		break
	}

	alias1 := h.Get
	alias2 := alias1
	alias2("TT") // want `use canonical header key "Tt" instead of "TT"`

	var byVar = h.Get
	byVarAlias := byVar
	byVarAlias("TT") // want `use canonical header key "Tt" instead of "TT"`

	alias3 := h.Get
	alias4 := alias3
	alias4 = strings.ToUpper
	alias4("TT")

	base := h.Get
	base, alias5 := strings.ToUpper, base
	alias5("TT") // want `use canonical header key "Tt" instead of "TT"`

	alias6 := h.Get
	useAlias := func() {
		alias6("TT") // want `use canonical header key "Tt" instead of "TT"`
	}
	useAlias()
}
