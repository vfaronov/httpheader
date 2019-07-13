/*
Package httpheader parses and generates standard HTTP headers.

For each header Foo-Bar covered by this package, there is a function
FooBar to parse it and SetFooBar (and sometimes AddFooBar) to generate it.

FooBar parses all valid Foo-Bar headers, but many invalid headers are also
tolerated and parsed to some extent. FooBar never errors, instead returning
whatever it can easily salvage. Do not assume that strings returned by FooBar
conform to the grammar of the protocol.

Likewise, SetFooBar doesn't validate parameter names or other tokens you supply.
However, it will automatically quote and escape your text where the grammar
admits an arbitrary quoted string or comment (RFC 7230 Section 3.2.6), such as
in parameter values.
*/
package httpheader
