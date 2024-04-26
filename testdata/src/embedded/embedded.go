package embedded

import "net/http"

type embeded struct {
	http.Header
}

func _() {
	embeded{}.Get("TT") // want `non-canonical header "TT", instead use: "Tt"`
}
