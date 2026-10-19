package expressions

import (
	"net/http"

	"expressions/keys"
)

const prefix = "X-"
const header = prefix + "HEADER"

func _(h http.Header, dynamic string) {
	h.Get("X-" + "HEADER")             // want `use "X-Header" instead of "X-HEADER"`
	h.Set(prefix+"HEADER", "value")    // want `use "X-Header" instead of "X-HEADER"`
	h.Get(keys.Header)                 // want `use "X-Header" instead of "X-HEADER"`
	h.Get((header))                    // want `use "X-Header" instead of "X-HEADER"`
	h.Get(string(keys.String(header))) // want `use "X-Header" instead of "X-HEADER"`
	h.Get(string('x'))                 // want `use "X" instead of "x"`
	h.Get(string(byte(120)))           // want `use "X" instead of "x"`
	h.Get("e" + "tag")                 // want `use "ETag" instead of "etag"`

	h.Get(("X-HEADER"))                    // want `use "X-Header" instead of "X-HEADER"`
	h.Get(string(keys.String("X-HEADER"))) // want `use "X-Header" instead of "X-HEADER"`
	h.Get(string(keys.String(`X-HEADER`))) // want `use "X-Header" instead of "X-HEADER"`
	h.Get("\x58-HEADER")                   // want `use "X-Header" instead of "X-HEADER"`

	h.Get("X-" + "Header")
	h.Get("E" + "Tag")
	h.Get(dynamic)
	h.Get(prefix + dynamic)
	h.Get(string(keys.String(dynamic)))
	h.Get(echo("X-HEADER"))
	h.Get(string(keys.String(echo("X-HEADER"))))
	h.Get("\xff")
}

func echo(key string) string { return key }
