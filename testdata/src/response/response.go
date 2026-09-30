package response

import "net/http"

func _(w http.ResponseWriter) {
	w.Header().Set("Test-HEader", "value") // want `use "Test-Header" instead of "Test-HEader"`
}
