package exclusions

import "net/http"

func _() {
	h := http.Header{}
	h.Add("Exclusion", "") // want `use canonical header key "exclusioN" instead of "Exclusion"`
	h.Add("exclusioN", "")
}
