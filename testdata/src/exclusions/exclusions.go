package exclusions

import "net/http"

func _() {
	h := http.Header{}
	h.Add("Exclusion", "") // want `use "exclusioN" instead of "Exclusion"`
	h.Add("exclusioN", "")
	h.Get("ex" + "clusion") // want `use "exclusioN" instead of "exclusion"`
	h.Get("ex" + "clusioN")
}
