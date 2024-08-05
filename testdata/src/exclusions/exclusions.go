package exclusions

import "net/http"

func _() {
	h := http.Header{}
	h.Add("Exclusion", "") // want `exclusion header "Exclusion", instead use: "exclusioN"`
	h.Add("exclusioN", "")
}
