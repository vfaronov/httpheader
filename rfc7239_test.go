package httpheader

import (
	"fmt"
	"net"
	"net/http"
	"testing"
)

func ExampleForwarded() {
	header := http.Header{"Forwarded": {
		`for=192.0.2.61;by="[2001:db8::fa40]";proto=https`,
		`for="[2001:db8::fa40]:19950";proto=http`,
	}}
	for _, elem := range Forwarded(header) {
		fmt.Println(elem.For.IP)
	}
	// Output: 192.0.2.61
	// 2001:db8::fa40
}

func ExampleAddForwarded() {
	header := http.Header{}
	AddForwarded(header, ForwardedElem{
		For:   Node{IP: net.IPv4(203, 0, 113, 195)},
		Proto: "https",
	})
	fmt.Print(header)
	// Output: map[Forwarded:[for=203.0.113.195;proto=https]]
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
			[]ForwardedElem{{For: Node{ObfuscatedNode: "_a"}}},
		},
		{
			http.Header{"Forwarded": {"for=_a;;;by=_b"}},
			[]ForwardedElem{
				{For: Node{ObfuscatedNode: "_a"}, By: Node{ObfuscatedNode: "_b"}},
			},
		},
		{
			http.Header{"Forwarded": {"for=_a;,;by=_b"}},
			[]ForwardedElem{
				{For: Node{ObfuscatedNode: "_a"}},
				{By: Node{ObfuscatedNode: "_b"}},
			},
		},
		{
			http.Header{"Forwarded": {
				`;For="[2001:db8:cafe::17]:4711";by=_Je8vvbnk5wmn;proto=HTTPS`,
			}},
			[]ForwardedElem{
				{
					For:   Node{IP: mustParseIP("2001:db8:cafe::17"), Port: 4711},
					By:    Node{ObfuscatedNode: "_Je8vvbnk5wmn"},
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
					For: Node{IP: mustParseIP("2001:db8:cafe::17"), Port: 4711},
					Ext: map[string]string{"qux": "90121"},
				},
				{
					By:    Node{ObfuscatedNode: "_Je8vvbnk5wmn"},
					Proto: "https",
				},
				{
					Host: "example.com",
				},
			},
		},
		{
			http.Header{"Forwarded": {`for="unknown"`}},
			[]ForwardedElem{{}},
		},
		{
			http.Header{"Forwarded": {`for=_Rrfqew8`}},
			[]ForwardedElem{{For: Node{ObfuscatedNode: "_Rrfqew8"}}},
		},
		{
			http.Header{"Forwarded": {`for="203.0.113.10"`}},
			[]ForwardedElem{{For: Node{IP: net.IPv4(203, 0, 113, 10)}}},
		},
		{
			http.Header{"Forwarded": {`for="[2001:db8:ae0::55]"`}},
			[]ForwardedElem{{For: Node{IP: mustParseIP("2001:db8:ae0::55")}}},
		},
		{
			http.Header{"Forwarded": {`for="unknown:28021"`}},
			[]ForwardedElem{{For: Node{Port: 28021}}},
		},
		{
			http.Header{"Forwarded": {`for="203.0.113.8:8080"`}},
			[]ForwardedElem{{For: Node{IP: net.IPv4(203, 0, 113, 8), Port: 8080}}},
		},
		{
			http.Header{"Forwarded": {`for="[2001:db8:4ca::20]:5033"`}},
			[]ForwardedElem{
				{For: Node{IP: mustParseIP("2001:db8:4ca::20"), Port: 5033}},
			},
		},
		{
			http.Header{"Forwarded": {`for="[2001:db8:4ca::20]:5"`}},
			[]ForwardedElem{
				{For: Node{IP: mustParseIP("2001:db8:4ca::20"), Port: 5}},
			},
		},
		{
			http.Header{"Forwarded": {`by="Unknown:1234"`}},
			[]ForwardedElem{{By: Node{Port: 1234}}},
		},
		{
			http.Header{"Forwarded": {`for="203.0.113.8:_ghu2"`}},
			[]ForwardedElem{
				{For: Node{IP: net.IPv4(203, 0, 113, 8), ObfuscatedPort: "_ghu2"}},
			},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Forwarded": {""}},
			nil,
		},
		{
			http.Header{"Forwarded": {"for=_a; by=_b, for=_c"}},
			[]ForwardedElem{
				{For: Node{ObfuscatedNode: "_a"}, By: Node{ObfuscatedNode: "_b"}},
				{For: Node{ObfuscatedNode: "_c"}},
			},
		},
		{
			http.Header{"Forwarded": {"for = _a;by = _b"}},
			[]ForwardedElem{
				{For: Node{ObfuscatedNode: "_a"}, By: Node{ObfuscatedNode: "_b"}},
			},
		},
		{
			http.Header{"Forwarded": {`for = "_a";by = "_b"`}},
			[]ForwardedElem{
				{For: Node{ObfuscatedNode: "_a"}, By: Node{ObfuscatedNode: "_b"}},
			},
		},
		{
			http.Header{"Forwarded": {`for=_a;by=", for=_b`}},
			[]ForwardedElem{
				{
					For: Node{ObfuscatedNode: "_a"},
					By:  Node{ObfuscatedNode: ", for=_b"},
				},
			},
		},
		{
			http.Header{"Forwarded": {`for=_a;by=", for="_b"`}},
			[]ForwardedElem{
				{
					For: Node{ObfuscatedNode: "_a"},
					By:  Node{ObfuscatedNode: ", for="},
					Ext: map[string]string{`_b"`: ""},
				},
			},
		},
		{
			http.Header{"Forwarded": {`for=_a;by="\, for=_b`}},
			[]ForwardedElem{
				{
					For: Node{ObfuscatedNode: "_a"},
					By:  Node{ObfuscatedNode: ", for=_b"},
				},
			},
		},
		{
			http.Header{"Forwarded": {
				`, For=" ";By=" ";Qux=" ";Host=" ";Proto=" "`,
			}},
			[]ForwardedElem{
				{
					For:   Node{ObfuscatedNode: " "},
					By:    Node{ObfuscatedNode: " "},
					Host:  " ",
					Proto: " ",
					Ext:   map[string]string{"qux": " "},
				},
			},
		},
		{
			http.Header{"Forwarded": {`for=_a by=_b, for=_c`}},
			[]ForwardedElem{
				{For: Node{ObfuscatedNode: "_a"}, By: Node{ObfuscatedNode: "_b"}},
				{For: Node{ObfuscatedNode: "_c"}},
			},
		},
		{
			http.Header{"Forwarded": {`for=_a=_b, for=_c`}},
			[]ForwardedElem{
				{For: Node{ObfuscatedNode: "_a"}},
				{For: Node{ObfuscatedNode: "_c"}},
			},
		},
		{
			http.Header{"Forwarded": {`for;by;qux, for=_c`}},
			[]ForwardedElem{
				{Ext: map[string]string{"qux": ""}},
				{For: Node{ObfuscatedNode: "_c"}},
			},
		},
		{
			http.Header{"Forwarded": {`for;=qux, for=_c`}},
			[]ForwardedElem{{}, {For: Node{ObfuscatedNode: "_c"}}},
		},
		{
			http.Header{"Forwarded": {`for="_a;\"_b\"";by="unknown:_c"`}},
			[]ForwardedElem{
				{
					For: Node{ObfuscatedNode: `_a;"_b"`},
					By:  Node{ObfuscatedPort: "_c"},
				},
			},
		},
		{
			http.Header{"Forwarded": {`for=""`}},
			[]ForwardedElem{{}},
		},
		{
			http.Header{"Forwarded": {`for=______:22602`}},
			[]ForwardedElem{{For: Node{ObfuscatedNode: "______", Port: 22602}}},
		},
		{
			http.Header{"Forwarded": {`for=[2001:db8:4ca::20]:5;by=_a`}},
			[]ForwardedElem{
				{
					For: Node{IP: mustParseIP("2001:db8:4ca::20"), Port: 5},
					By:  Node{ObfuscatedNode: "_a"},
				},
			},
		},
		{
			http.Header{"Forwarded": {`for="2001:db8:ae0::55"`}},
			[]ForwardedElem{
				{For: Node{ObfuscatedNode: "2001:db8:ae0:", Port: 55}},
			},
		},
		{
			http.Header{"Forwarded": {`by=":1309"`}},
			[]ForwardedElem{{By: Node{Port: 1309}}},
		},
		{
			http.Header{"Forwarded": {`for="[2001:db8:ae0::55]:"`}},
			[]ForwardedElem{{For: Node{IP: mustParseIP("2001:db8:ae0::55")}}},
		},
		{
			http.Header{"Forwarded": {`for=":"`}},
			[]ForwardedElem{{}},
		},
		{
			http.Header{"Forwarded": {`for="203.0.113.8:"`}},
			[]ForwardedElem{{For: Node{IP: net.IPv4(203, 0, 113, 8)}}},
		},
		{
			http.Header{"Forwarded": {`for="[2001:db8:ae0::55"`}},
			[]ForwardedElem{
				{For: Node{ObfuscatedNode: "2001:db8:ae0:", Port: 55}},
			},
		},
		{
			http.Header{"Forwarded": {`by="2001:db8:ae0::55]"`}},
			[]ForwardedElem{{By: Node{IP: mustParseIP("2001:db8:ae0::55")}}},
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
			[]ForwardedElem{},
			http.Header{},
		},
		{
			[]ForwardedElem{
				{
					For:   Node{ObfuscatedNode: "_a"},
					By:    Node{IP: mustParseIP("2001:db8:ae0::55")},
					Proto: "ABCDP",
					Ext:   map[string]string{"foo": ""},
				},
			},
			http.Header{
				"Forwarded": {`for=_a;by="[2001:db8:ae0::55]";proto=ABCDP;foo=""`},
			},
		},
		{
			[]ForwardedElem{{}, {}},
			http.Header{"Forwarded": {"for=unknown, for=unknown"}},
		},
		{
			[]ForwardedElem{{By: Node{Port: 8080}}, {}},
			http.Header{"Forwarded": {`by="unknown:8080", for=unknown`}},
		},
		{
			[]ForwardedElem{
				{
					For: Node{
						IP:             net.IPv4(203, 0, 113, 70),
						ObfuscatedNode: "_EF99AC",
					},
				},
			},
			http.Header{"Forwarded": {"for=203.0.113.70"}},
		},
		{
			[]ForwardedElem{
				{For: Node{IP: net.IPv4(203, 0, 113, 70), Port: 44831}},
			},
			http.Header{"Forwarded": {`for="203.0.113.70:44831"`}},
		},
		{
			[]ForwardedElem{
				{For: Node{Port: 44831, ObfuscatedPort: "_a"}},
			},
			http.Header{"Forwarded": {`for="unknown:44831"`}},
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
		[]ForwardedElem{
			{
				For: Node{
					IP:             net.IP{},
					Port:           0,
					ObfuscatedNode: "empty",
					ObfuscatedPort: "_obfID | empty",
				},
				By: Node{
					IP:             nil,
					Port:           9999,
					ObfuscatedNode: "_obfID | empty",
					ObfuscatedPort: "empty",
				},
				Host:  "token | empty",
				Proto: "lower token | empty",
				Ext:   map[string]string{"lower token": "quotable | empty"},
			},
		},
	)
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

func mustParseIP(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		panic(fmt.Sprintf("cannot parse IP: %q", s))
	}
	return ip
}
