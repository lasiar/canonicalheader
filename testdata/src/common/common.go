package common

import (
	"net/http"
)

const constTestHeader = "testHeaderValue"

func p() {
	v := http.Header{}
	v.Get(constTestHeader) // want `use "Testheadervalue" instead of "testHeaderValue"`

	v.Get("Test-HEader")           // want `use "Test-Header" instead of "Test-HEader"`
	v.Set("Test-HEader", "value")  // want `use "Test-Header" instead of "Test-HEader"`
	v.Add("Test-HEader", "value")  // want `use "Test-Header" instead of "Test-HEader"`
	v.Del("Test-HEader")           // want `use "Test-Header" instead of "Test-HEader"`
	v.Values("Test-HEader")        // want `use "Test-Header" instead of "Test-HEader"`
	v.Values(`Raw-STRING-Literal`) // want `use "Raw-String-Literal" instead of "Raw-STRING-Literal"`

	v.Set("Test-Header", "value")
	v.Add("Test-Header", "value")
	v.Del("Test-Header")
	v.Values("Test-Header")

	var someString = ""
	v.Get(someString)

	v.Write(nil)
	v.Clone()
	v.WriteSubset(nil, nil)
}
