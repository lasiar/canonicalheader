package mapaccess

import "net/http"

const constTestHeader = "testHeaderValue"

type myHeader = http.Header

func _(w http.ResponseWriter, r *http.Request) {
	h := http.Header{}

	_ = h["Test-HEader"]           // want `use "Test-Header" instead of "Test-HEader"`
	_, _ = h[`Raw-STRING-Literal`] // want `use "Raw-String-Literal" instead of "Raw-STRING-Literal"`
	_ = h[constTestHeader]         // want `use "Testheadervalue" instead of "testHeaderValue"`
	h["Test-HEader"] = nil         // want `use "Test-Header" instead of "Test-HEader"`
	delete(h, "Test-HEader")       // want `use "Test-Header" instead of "Test-HEader"`

	h["X-HEader"] = append(h["X-HEader"], "value") // want `use "X-Header" instead of "X-HEader"` `use "X-Header" instead of "X-HEader"`

	// Exclusions do not apply: net/http stores "Etag", so "ETag" never matches.
	_ = h["ETag"] // want `use "Etag" instead of "ETag"`

	_ = w.Header()["Test-HEader"] // want `use "Test-Header" instead of "Test-HEader"`
	_ = r.Header["Test-HEader"]   // want `use "Test-Header" instead of "Test-HEader"`
	p := &h
	_ = (*p)["Test-HEader"]       // want `use "Test-Header" instead of "Test-HEader"`
	_ = myHeader{}["Test-HEader"] // want `use "Test-Header" instead of "Test-HEader"`

	_ = http.Header{
		"Test-HEader": nil, // want `use "Test-Header" instead of "Test-HEader"`
		"Test-Header": nil,
	}
	_ = map[string]http.Header{
		"not-a-header": {"Test-HEader": nil}, // want `use "Test-Header" instead of "Test-HEader"`
	}

	_ = h["Test-Header"]
	h["Test-Header"] = nil
	delete(h, "Test-Header")

	var key string
	_ = h[key]
	delete(h, key)

	m := map[string][]string{}
	_ = m["not-a-header"]
	delete(m, "not-a-header")
}
