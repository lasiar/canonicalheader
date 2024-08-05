package _struct

import "net/http"

type HeaderStruct struct {
	header http.Header
}

func (h HeaderStruct) _() {
	h.header.Get("TT") // want `use "Tt" instead of "TT"`
}

func _() {
	HeaderStruct{}.header.Get("TT") // want `use "Tt" instead of "TT"`
}
