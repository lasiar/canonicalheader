package assigned

import (
	"fmt"
	"net/http"
)

func _() {
	h := http.Header{}

	i, g := 0, h.Del
	fmt.Println(i)
	g("TT") // want `use "Tt" instead of "TT"`

	f := h.Get
	f("TT") // want `use "Tt" instead of "TT"`
}
