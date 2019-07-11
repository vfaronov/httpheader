/*
Package httpheader parses and generates standard HTTP headers.

For each header Foo-Bar covered by this package, there is a function
FooBar to parse it and SetFooBar (and sometimes AddFooBar) to generate it.

FooBar parses all valid Foo-Bar headers (though see Bugs), but many invalid
headers are also tolerated and parsed to some extent. Do not assume that strings
returned by FooBar conform to the grammar of the protocol. Sanitize them before
including them in HTML or other output.

Likewise, SetFooBar doesn't validate parameter names or other tokens you supply.
However, it will automatically quote and escape your text where the grammar
admits a quoted string or comment (RFC 7230 Section 3.2.6), such as in parameter
values.
*/
package httpheader
