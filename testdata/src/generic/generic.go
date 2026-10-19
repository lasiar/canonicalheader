package generic

import "net/http"

type value[T any] struct{}

func (value[T]) Get(string) T {
	var zero T
	return zero
}

func (*value[T]) Set(string, T) {}

type headers[T any] struct {
	http.Header
}

func _() {
	h := http.Header{}
	h.Get("X-HEADER") // want `use "X-Header" instead of "X-HEADER"`

	v := value[string]{}
	v.Get("X-HEADER")
	v.Set("X-HEADER", "value")

	(v.Get)("X-HEADER")
	(v.Set)("X-HEADER", "value")
	(value[string].Get)(v, "X-HEADER")
	((*value[string]).Set)(&v, "X-HEADER", "value")

	get := v.Get
	get("X-HEADER")
	set := v.Set
	set("X-HEADER", "value")

	h.Set("X-HEADER", "value") // want `use "X-Header" instead of "X-HEADER"`
	headerGet := h.Get
	headerGet("X-HEADER") // want `use "X-Header" instead of "X-HEADER"`
	h.Get("X-Header")

	embedded := headers[string]{Header: h}
	embedded.Get("X-HEADER") // want `use "X-Header" instead of "X-HEADER"`
}
