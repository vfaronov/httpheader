package httpheader

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"reflect"
	"testing"
)

func ExampleForwarded() {
	header := http.Header{"Forwarded": {
		`for=192.0.2.61;by="[2001:db8::fa40]";proto=https`,
		`for="[2001:db8::fa40]:19950";proto=http`,
	}}
	for _, elem := range Forwarded(header) {
		fmt.Println(elem.ForAddr())
	}
	// Output: 192.0.2.61 0
	// 2001:db8::fa40 19950
}

func ExampleForwarded_malformed() {
	// An attacker submits a malformed Forwarded value
	// (note the unterminated quoted string), hoping to
	// impersonate 10.5.7.13 by "shadowing" later values.
	v := `for=10.5.7.13;by="`

	// A gateway blindly appends the attacker's true IP address.
	v += `, for=198.51.100.73;by=10.14.2.19`

	// Both elements are parsed successfully.
	header := http.Header{"Forwarded": {v}}
	fmt.Printf("%+v", Forwarded(header))

	// However, none of this information can be trusted
	// unless you know (perhaps from configuration) that
	// the last element is always appended by a trusted gateway.

	// Output: [{By: For:10.5.7.13 Host: Proto: Ext:map[]} {By:10.14.2.19 For:198.51.100.73 Host: Proto: Ext:map[]}]
}

func ExampleAddForwarded() {
	header := http.Header{}
	AddForwarded(header, ForwardedElem{
		For:   "[2001:db8:cafe::17]",
		Proto: "https",
	})
	fmt.Println(header)
	// Output: map[Forwarded:[for="[2001:db8:cafe::17]";proto=https]]
}

func TestForwarded(t *testing.T) {
	tests := []struct {
		header http.Header
		result []ForwardedElem
	}{
		// Valid headers.
		{
			http.Header{"Forwarded": {";"}},
			[]ForwardedElem{{}},
		},
		{
			http.Header{"Forwarded": {"for=_a;"}},
			[]ForwardedElem{{For: "_a"}},
		},
		{
			http.Header{"Forwarded": {`for="_a;\"_b\"";`}},
			[]ForwardedElem{{For: `_a;"_b"`}},
		},
		{
			http.Header{"Forwarded": {"for=_a;;;by=_b"}},
			[]ForwardedElem{{For: "_a", By: "_b"}},
		},
		{
			http.Header{"Forwarded": {"for=_a;,;by=_b"}},
			[]ForwardedElem{{For: "_a"}, {By: "_b"}},
		},
		{
			http.Header{"Forwarded": {
				`;For="[2001:db8:cafe::17]:4711";by=_Je8vvbnk5wmn;proto=HTTPS`,
			}},
			[]ForwardedElem{
				{
					For:   "[2001:db8:cafe::17]:4711",
					By:    "_Je8vvbnk5wmn",
					Proto: "https",
				},
			},
		},
		{
			http.Header{"Forwarded": {
				`;For="[2001:db8:cafe::17]:4711";Qux=90121`,
				`;by=_Je8vvbnk5wmn;proto=HTTPS,host=example.com`,
			}},
			[]ForwardedElem{
				{
					For: "[2001:db8:cafe::17]:4711",
					Ext: map[string]string{"qux": "90121"},
				},
				{
					By:    "_Je8vvbnk5wmn",
					Proto: "https",
				},
				{
					Host: "example.com",
				},
			},
		},

		// Invalid headers. Except as documented on the Forwarded function,
		// precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Forwarded": {""}},
			nil,
		},
		{
			http.Header{"Forwarded": {"for=_a; by=_b, for=_c"}},
			[]ForwardedElem{{For: "_a"}, {For: "_c"}},
		},
		{
			http.Header{"Forwarded": {"for = _a;by = _b"}},
			[]ForwardedElem{{}},
		},
		{
			http.Header{"Forwarded": {`for = "_a";by = "_b"`}},
			[]ForwardedElem{{}},
		},
		{
			http.Header{"Forwarded": {`for=_a;by=", for=_b`}},
			[]ForwardedElem{{For: "_a"}, {For: "_b"}},
		},
		{
			http.Header{"Forwarded": {`for=_a;by="\, for=_b`}},
			[]ForwardedElem{{For: "_a"}, {For: "_b"}},
		},
		{
			http.Header{"Forwarded": {
				`, For=" ";By=" ";Qux=" ";Host=" ";Proto=" "`,
			}},
			[]ForwardedElem{
				{
					For:   " ",
					By:    " ",
					Host:  " ",
					Proto: " ",
					Ext:   map[string]string{"qux": " "},
				},
			},
		},
		{
			http.Header{"Forwarded": {`for=_a by=_b, for=_c`}},
			[]ForwardedElem{{For: "_a"}, {For: "_c"}},
		},
		{
			http.Header{"Forwarded": {`for=_a=_b, for=_c`}},
			[]ForwardedElem{{For: "_a"}, {For: "_c"}},
		},
		{
			http.Header{"Forwarded": {`for;by;qux, for=_c`}},
			[]ForwardedElem{{}, {For: "_c"}},
		},
		{
			http.Header{"Forwarded": {`for;=qux, for=_c`}},
			[]ForwardedElem{{}, {For: "_c"}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Forwarded(test.header))
		})
	}
}

func TestSetForwarded(t *testing.T) {
	tests := []struct {
		input  []ForwardedElem
		result http.Header
	}{
		{
			[]ForwardedElem{
				{
					For:   "_a",
					By:    "[2001:db8:ae0::55]",
					Proto: "ABCDP",
					Ext:   map[string]string{"foo": ""},
				},
			},
			http.Header{
				"Forwarded": {`for=_a;by="[2001:db8:ae0::55]";proto=ABCDP;foo=""`},
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetForwarded(header, test.input)
			checkGenerate(t, test.input, test.result, header)
		})
	}
}

func TestForwardedFuzz(t *testing.T) {
	checkFuzz(t, "Forwarded", Forwarded, SetForwarded)
}

func TestForwardedRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetForwarded, Forwarded,
		[]ForwardedElem{{
			For:   "quotable",
			By:    "quotable | empty",
			Host:  "quotable | empty",
			Proto: "lower quotable | empty",
			Ext:   map[string]string{"lower token": "quotable | empty"},
		}},
	)
}

func TestForwardedRecover(t *testing.T) {
	// Property-based test: A valid element at the end of a field-value
	// is recovered regardless of what precedes it.
	for i := 0; i < 100; i++ {
		t.Run("", func(t *testing.T) {
			r := rand.New(rand.NewSource(int64(i)))
			junk := make([]byte, r.Intn(64))
			r.Read(junk)
			header := http.Header{
				"Forwarded": {string(junk) + ", for=_a;by=_b"},
			}
			t.Logf("header: %q", header)
			parsed := Forwarded(header)
			t.Logf("parsed: %#v", parsed)
			if len(parsed) == 0 {
				t.Fatalf("no elements parsed")
			}
			expected := ForwardedElem{For: "_a", By: "_b"}
			actual := parsed[len(parsed)-1]
			if !reflect.DeepEqual(expected, actual) {
				t.Fatalf("failed to recover:\nexpected: %#v\nactual:   %#v",
					expected, actual)
			}
		})
	}
}

func BenchmarkForwarded(b *testing.B) {
	header := http.Header{"Forwarded": {
		`by=_proxyACe430bZq8g8d;for=10.41.132.82;acl="check passed"`,
		`by="[2001:db8:ae0::55]";for=198.51.100.67;proto=https;host=example.com`,
		`for="[2001:db8:ae0::55]";proto=http, for="[2001:db8:f6::c3]";proto=http`,
	}}
	for i := 0; i < b.N; i++ {
		Forwarded(header)
	}
}

func TestForwardedElemByAddr(t *testing.T) {
	tests := []struct {
		value string
		ip    net.IP
		port  int
	}{
		// Valid node values.
		{"", nil, 0},
		{"unknown", nil, 0},
		{"_Rrfqew8", nil, 0},
		{"203.0.113.10", net.IPv4(203, 0, 113, 10), 0},
		{"[2001:db8:ae0::55]", mustParseIP("2001:db8:ae0::55"), 0},
		{"unknown:8021", nil, 8021},
		{"______:22602", nil, 22602},
		{"203.0.113.8:8080", net.IPv4(203, 0, 113, 8), 8080},
		{"[2001:db8:4ca::20]:5033", mustParseIP("2001:db8:4ca::20"), 5033},
		{"[2001:db8:4ca::20]:5", mustParseIP("2001:db8:4ca::20"), 5},
		{"unknown:_1234", nil, 0},
		{"203.0.113.8:_ghu2", net.IPv4(203, 0, 113, 8), 0},

		// Invalid node values.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{"2001:db8:ae0::55", nil, 55},
		{":1309", nil, 1309},
		{"[2001:db8:4ca::20]:", mustParseIP("2001:db8:4ca::20"), 0},
		{":", nil, 0},
		{"203.0.113.8:", net.IPv4(203, 0, 113, 8), 0},
		{"[2001:db8:4ca::20", nil, 20},
		{"2001:db8:4ca::20]", mustParseIP("2001:db8:4ca::20"), 0},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			ip, port := ForwardedElem{By: test.value}.ByAddr()
			if !ip.Equal(test.ip) {
				t.Errorf("parsing %q\nexpected IP: %s\nactual IP:   %s",
					test.value, test.ip, ip)
			}
			if port != test.port {
				t.Errorf("parsing %q\nexpected port: %v\nactual port:   %v",
					test.value, test.port, port)
			}
		})
	}
}

func mustParseIP(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		panic(fmt.Sprintf("cannot parse IP: %q", s))
	}
	return ip
}
