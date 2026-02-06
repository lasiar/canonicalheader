.PHONY:

test:
	go test -v -race ./...

fmt:
	golangci-lint fmt

fmt-check:
	golangci-lint fmt -d

lint: fmt-check
	golangci-lint -v run ./...

linter: lint

generate:
	go run ./cmd/initialismer/main.go -target="mapping" > ./initialism.go
	go run ./cmd/initialismer/main.go -target="test" > ./testdata/src/initialism/initialism.go
	go run ./cmd/initialismer/main.go -target="test-golden" > ./testdata/src/initialism/initialism.go.golden
	gofmt -w ./initialism.go ./testdata/src/initialism/initialism.go ./testdata/src/initialism/initialism.go.golden
