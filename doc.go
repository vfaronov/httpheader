/*
Package httpheader parses and generates standard HTTP headers.

For each header Foo-Bar covered by this package, there is a function FooBar
to parse it, SetFooBar to generate it, and sometimes AddFooBar to append to it.
Note that AddFooBar creates a new Foo-Bar header line (like http.Header's
Add method), which, although correct, is poorly supported by some recipients.

FooBar parses all valid Foo-Bar headers, but many invalid headers are also
tolerated and parsed to some extent. FooBar never errors, instead returning
whatever it can easily salvage. Do not assume that strings returned by FooBar
conform to the grammar of the protocol.

Likewise, SetFooBar doesn't validate parameter names or other tokens you supply.
However, it will automatically quote and escape your text where the grammar
admits an arbitrary quoted string or comment (RFC 7230 Section 3.2.6), such as
in parameter values.

Tokens that are known to be case-insensitive, like directive or parameter names,
are lowercased by FooBar. Any slices or maps returned by FooBar may be nil
when there is no corresponding input data.
*/
package httpheader
