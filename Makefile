.PHONY: check checkfmt fmt vet test coverhtml example

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

example:
# Replace the example in README.md with tested code.
	sed '/^## Example/,$$d' README.md >new.README.md
	echo '## Example' >>new.README.md
	echo '' >>new.README.md
	sed -n 's/^/\t/; /const request/,$$p' example_test.go >>new.README.md
	mv -f new.README.md README.md
