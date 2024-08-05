package alias

import "net/http"

type myHeader = http.Header

func _() {
	myHeader{}.Get("TT") // want `use "Tt" instead of "TT"`
}
