.PHONY: check checkfmt fmt vet test coverhtml

check: checkfmt vet test

checkfmt:
	test -z "$$(find . -type f -name '*.go' -exec gofmt -s -d {} \;)"

fmt:
	find . -type f -name '*.go' -exec gofmt -s -w {} \;

vet:
	go vet .

test:
	go test -cover .

coverhtml:
	go test -coverprofile=cover.out .
	go tool cover -html=cover.out
