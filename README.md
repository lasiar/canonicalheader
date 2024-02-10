# funcliner

Golang linter for check func params on one line or each params.

### Example

before

```go
package main

func f(
	p1, p2 int,
	p3 bool,
) {}

```

after

```go
package main

func f(
	p1,
	p2 int,
	p3 bool,
) {}

```
