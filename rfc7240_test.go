package httpheader

import (
	"fmt"
	"net/http"
)

func ExamplePrefer() {
	header := http.Header{"Prefer": []string{
		"wait=10, respond-async",
		`Foo; Bar="quux, xyzzy"`,
	}}
	prefer := Prefer(header)
	fmt.Printf("%q\n", prefer["wait"])
	fmt.Printf("%q\n", prefer["respond-async"])
	fmt.Printf("%q\n", prefer["foo"])
	// Output: {"10" map[]}
	// {"" map[]}
	// {"" map["bar":"quux, xyzzy"]}
}

func ExampleAddPrefer() {
	header := http.Header{}
	SetPrefer(header, map[string]Pref{
		"wait":          {"10", nil},
		"respond-async": {},
		"foo": {
			Value:  "bar, baz",
			Params: map[string]string{"quux": "xyzzy"},
		},
	})
}
