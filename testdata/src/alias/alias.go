package alias

import "net/http"

type myHeader = http.Header

func _() {
	myHeader{}.Get("TT") // want `use canonical header key "Tt" instead of "TT"`
}
