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
	sed '/^## Example/,$$d' README.md >README.new.md
	echo '## Example' >>README.new.md
	echo '' >>README.new.md
	sed -n '/^}/d; s/^\t//; s/^/\t/; /const/,$$p' example_test.go >>README.new.md
	mv -f README.new.md README.md
