.PHONY: test lint qa coverhtml fmt example

test:
	go test -cover .

lint:
	golangci-lint run

qa: test lint

coverhtml:
	go test -coverprofile=cover.out .
	go tool cover -html=cover.out

fmt:
	find . -type f -name '*.go' -exec gofmt -s -w {} \;

example:
# Replace the example in README.md with tested code.
	sed '/^## Example/,$$d' README.md >new.README.md
	echo '## Example' >>new.README.md
	echo '' >>new.README.md
	sed -n 's/^/\t/; /const request/,$$p' example_test.go >>new.README.md
	mv -f new.README.md README.md
