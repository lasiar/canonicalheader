package p

import "context"

func f1(
	someString string,
	someInt int,
	someBool bool,
) (string, error) {
	return "", nil
}

func f2(someString string, // want `f2 params on one line or each parameter on a separate line`
	someInt int,
	someBool bool,
) (string, int, error) {
	return "", 0, nil
}

func f3(someString string, someInt int, // want `f3 params on one line or each parameter on a separate line`
	someBool bool,
) (string, int, error) {
	return "", 0, nil
}

func f4(someString string, someInt int, someBool bool) (string, int, error) {
	return "", 0, nil
}

func f5( // want `f5 params on one line or each parameter on a separate line`
	someString string,
	someInt int,
	someBool bool) (string, int, error) {
	return "", 0, nil
}

func f6( // want `f6 params on one line or each parameter on a separate line`
	someString, someInt int,
	someBool bool,
) (string, int, error) {
	return "", 0, nil
}

func f7(v1, v2, v3, v4, v5 string) {}

type t struct{}

func (r t) f7(
	ctx context.Context,
	i,
	i1 int64,
) {
}

func empty() {}
