/*
Package httpheader parses and serializes HTTP header values according to standards.

For each header Foo-Bar known to this package, there is a function
FooBar to parse it and SetFooBar (and sometimes AddFooBar) to serialize it.

FooBar parses all valid Foo-Bar headers (but see Bugs). Many invalid headers are
also tolerated and parsed to some extent, but no guarantees are made about them.
Because of this laxity, values returned by FooBar may not conform to the grammar
of the protocol. For example, a string that should be a simple token might
erroneously contain a double quote. If this is important for you (like if you are
going to put this string into HTML), sanitize it.

Likewise, SetFooBar performs no validation. If you supply a string with whitespace
where the grammar of Foo-Bar only admits one token, such as in a parameter name,
the resulting Foo-Bar value will be malformed. On the other hand, wherever
the grammar admits a quoted string or comment (RFC 7230 Section 3.2.6), such as
in parameter values, the text you supply will automatically be quoted and escaped
as necessary, so it may contain any bytes except control characters.
*/
package httpheader
