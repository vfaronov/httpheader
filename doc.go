/*
Package httpheader parses and serializes HTTP header values according to standards.

For each header Foo-Bar known to this package, there is a function
FooBar to parse it and SetFooBar (and sometimes AddFooBar) to serialize it.

FooBar parses all valid Foo-Bar headers. Many invalid headers are also tolerated
and parsed to some extent, but no guarantees are made about them.

Likewise, SetFooBar performs no validation. For example, if you supply a string
with whitespace where the grammar of Foo-Bar only admits one token, such as in a
parameter name, the resulting Foo-Bar value will be malformed. On the other hand,
wherever the grammar admits a quoted string or comment (RFC 7230 Section 3.2.6),
such as in most parameter values, the text you supply will automatically be quoted
and escaped as necessary, so it may contain any bytes except control characters.
*/
package httpheader
