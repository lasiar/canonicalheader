package underlying

import "net/http"

type myString string

const c = ""

func _() {
	http.Header{}.Get(st(c))
	http.Header{}.Get(st("hello-world"))
	http.Header{}.Get(string(string(myString("TT")))) // want `use "Tt" instead of "TT"`
	http.Header{}.Get(t())
}

func t() string { return "" }

func st(str string) string { return str }
