/*
Package httpheader parses and serializes HTTP headers according to standards.

For each header Foo-Bar known to this package, there is a function
FooBar to parse it and SetFooBar (and sometimes AddFooBar) to serialize it.
FooBar parses all valid Foo-Bar headers. Many invalid headers are also tolerated
and parsed to some extent, but no guarantees are made about them.
*/
package httpheader
