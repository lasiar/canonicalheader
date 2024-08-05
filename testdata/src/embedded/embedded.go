package embedded

import "net/http"

type embedded struct {
	http.Header
}

func _() {
	embedded{}.Get("TT") // want `use "Tt" instead of "TT"`
}
