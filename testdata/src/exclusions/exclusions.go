package exclusions

import "net/http"

func _() {
	h := http.Header{}
	h.Add("Exclusion", "") // want `use "exclusioN" instead of "Exclusion"`
	h.Add("exclusioN", "")
}
