package responsewriter

import "net/http"

// No explicit http.Header type reference: Header is an interface method here.
func _(w http.ResponseWriter) {
	w.Header().Get("X-HEADER")          // want `use "X-Header" instead of "X-HEADER"`
	w.Header().Set("X-HEADER", "value") // want `use "X-Header" instead of "X-HEADER"`
	w.Header().Add("X-HEADER", "value") // want `use "X-Header" instead of "X-HEADER"`
	w.Header().Del("X-HEADER")          // want `use "X-Header" instead of "X-HEADER"`
	w.Header().Values("X-HEADER")       // want `use "X-Header" instead of "X-HEADER"`
	w.Header().Get("X-Header")

	h := w.Header()
	h.Get("X-HEADER") // want `use "X-Header" instead of "X-HEADER"`
	get := h.Get
	get("X-HEADER") // want `use "X-Header" instead of "X-HEADER"`
}

type Header struct{}

func (Header) Get(string) {}

func _() {
	Header{}.Get("X-HEADER")
}
