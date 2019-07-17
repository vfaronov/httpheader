package httpheader

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func ExampleLink() {
	// In real code, base := resp.Request.URL
	base, _ := url.Parse("https://api.example/articles/123")
	header := http.Header{"Link": {
		`</articles/124>; rel="next"; title*=utf-8''Witaj%20%C5%9Bwiecie!`,
		`<./>;rel=up, <https://api.example/doc>;rel=help;title="API help"`,
	}}
	links := Link(header, base)
	for _, link := range links {
		fmt.Printf("%-5s %-35s %q\n", link.Rel, link.Target, link.Title)
	}
	// Output: next  https://api.example/articles/124    "Witaj świecie!"
	// up    https://api.example/articles/       ""
	// help  https://api.example/doc             "API help"
}

func ExampleAddLink() {
	header := http.Header{}
	AddLink(header, LinkElem{
		Target: &url.URL{Path: "/articles/124"},
		Rel:    "next",
		Title:  "Witaj świecie!",
	})
	header.Write(os.Stdout)
	// Output: Link: </articles/124>; rel=next; title*=UTF-8''Witaj%20%C5%9Bwiecie!
}

const testBase = "http://x.test/a"

func TestLink(t *testing.T) {
	tests := []struct {
		header http.Header
		result []LinkElem
	}{
		// Valid headers.
		{
			http.Header{"Link": {""}},
			[]LinkElem{},
		},
		{
			http.Header{"Link": {"<https://example.net/>;rel=up"}},
			[]LinkElem{{Rel: "up", Target: U("https://example.net/")}},
		},
		{
			http.Header{"Link": {"<../>;rel=up"}},
			[]LinkElem{{Rel: "up", Target: U("http://x.test/")}},
		},
		{
			http.Header{"Link": {"<>;\trel=self"}},
			// I thought this should resolve to http://x.test/, but probably
			// net/url is correct here, and I'm not going to override it anyway.
			[]LinkElem{{Rel: "self", Target: U("http://x.test/a")}},
		},
		{
			http.Header{"Link": {"<b>;\trel=next"}},
			[]LinkElem{{Rel: "next", Target: U("http://x.test/b")}},
		},
		{
			http.Header{"Link": {"<#self>;  rel = describes"}},
			[]LinkElem{
				{Rel: "describes", Target: U("http://x.test/a#self")},
			},
		},
		{
			http.Header{"Link": {`<urn:whatever:123>;rel="urn:whatever:456"`}},
			[]LinkElem{{Rel: "urn:whatever:456", Target: U("urn:whatever:123")}},
		},
		{
			http.Header{"Link": {`<b>; rel="next prefetch"; hreflang=en; extra`}},
			[]LinkElem{
				{
					Rel:      "next",
					Target:   U("http://x.test/b"),
					HrefLang: []string{"en"},
					Ext:      map[string]string{"extra": ""},
				},
				{
					Rel:      "prefetch",
					Target:   U("http://x.test/b"),
					HrefLang: []string{"en"},
					Ext:      map[string]string{"extra": ""},
				},
			},
		},
		{
			http.Header{"Link": {`<b>; rel="next  prefetch"`}},
			[]LinkElem{
				{Rel: "next", Target: U("http://x.test/b")},
				{Rel: "prefetch", Target: U("http://x.test/b")},
			},
		},
		{
			http.Header{"Link": {`<b>;rel="next chapter";rev="prev chapter"`}},
			[]LinkElem{
				{Rel: "next", Target: U("http://x.test/b")},
				{Rel: "chapter", Target: U("http://x.test/b")},
			},
		},
		{
			http.Header{"Link": {`<a.rdf>;anchor="#self";rel="DescribedBy"`}},
			[]LinkElem{
				{
					Anchor: U("http://x.test/a#self"),
					Rel:    "describedby",
					Target: U("http://x.test/a.rdf"),
				},
			},
		},
		{
			http.Header{"Link": {`</> ; rel = "https://vocab.example/memberOf"`}},
			[]LinkElem{
				{
					// "When extension relation types are compared, they MUST be
					// compared as strings [...] in a case-insensitive fashion"
					Rel:    "https://vocab.example/memberof",
					Target: U("http://x.test/"),
				},
			},
		},
		{
			http.Header{"Link": {`</edit>;rel=edit-form;anchor=""`}},
			[]LinkElem{
				{
					Anchor: U("http://x.test/a"),
					Rel:    "edit-form",
					Target: U("http://x.test/edit"),
				},
			},
		},
		{
			http.Header{"Link": {`<b>;rel=next;rel*=utf-8''prefetch`}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Ext:    map[string]string{"rel": "prefetch"},
				},
			},
		},
		{
			http.Header{"Link": {"<b>; rel=next;title=Hello,,"}},
			[]LinkElem{
				{Rel: "next", Target: U("http://x.test/b"), Title: "Hello"},
			},
		},
		{
			http.Header{"Link": {`<b>; rel=next;title="Hello, world",`}},
			[]LinkElem{
				{Rel: "next", Target: U("http://x.test/b"), Title: "Hello, world"},
			},
		},
		{
			http.Header{"Link": {
				`<b>; rel=next; title*=utf-8''Hell%C3%B5%2C%20w%C3%B5rld`,
			}},
			[]LinkElem{
				{Rel: "next", Target: U("http://x.test/b"), Title: "Hellõ, wõrld"},
			},
		},
		{
			http.Header{"Link": {
				`<b>; rel=next; title="Hello, world"; title*=utf-8''Hell%C3%B5%2C%20w%C3%B5rld`,
			}},
			[]LinkElem{
				{Rel: "next", Target: U("http://x.test/b"), Title: "Hellõ, wõrld"},
			},
		},
		{
			http.Header{"Link": {
				`<b>; rel=next; title*=utf-8''Hell%C3%B5%2C%20w%C3%B5rld; title="Hello, world"`,
			}},
			[]LinkElem{
				{Rel: "next", Target: U("http://x.test/b"), Title: "Hellõ, wõrld"},
			},
		},
		{
			http.Header{"Link": {"<b>;Rel=next;Qux"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Ext:    map[string]string{"qux": ""},
				},
			},
		},
		{
			http.Header{"Link": {"<b>;Rel=next;Qux=xyzzy"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Ext:    map[string]string{"qux": "xyzzy"},
				},
			},
		},
		{
			http.Header{"Link": {"<b>;Rel=next;Qux=xyzzy;Qux*=utf-8''xyzzy!"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Ext:    map[string]string{"qux": "xyzzy!"},
				},
			},
		},
		{
			http.Header{"Link": {"<b>;Rel=next;Qux*=utf-8''xyzzy!;Qux=xyzzy"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Ext:    map[string]string{"qux": "xyzzy!"},
				},
			},
		},
		{
			http.Header{"Link": {`, <b>;rel=next;hreflang=en-US, </>;rel=up;type="Text/HTML"`}},
			[]LinkElem{
				{
					Rel:      "next",
					Target:   U("http://x.test/b"),
					HrefLang: []string{"en-us"},
				},
				{
					Rel:    "up",
					Target: U("http://x.test/"),
					Type:   "text/html",
				},
			},
		},
		{
			http.Header{"Link": {
				`</TheBook/chapter2>; rel="previous"; title*=UTF-8'de'letztes%20Kapitel, </TheBook/chapter4>; rel="next"; title*=UTF-8'de'n%c3%a4chstes%20Kapitel`,
			}},
			[]LinkElem{
				{
					Rel:    "previous",
					Target: U("http://x.test/TheBook/chapter2"),
					Title:  "letztes Kapitel",
				},
				{
					Rel:    "next",
					Target: U("http://x.test/TheBook/chapter4"),
					Title:  "nächstes Kapitel",
				},
			},
		},
		{
			http.Header{"Link": {"<b;qux?xyzzy=>;rel=next"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b;qux?xyzzy="),
				},
			},
		},
		{
			http.Header{"Link": {`<https://example.com/privacy>; anchor="https://example.com/"; rel=privacy-policy; title="Privacy"; type="Application/XHTML+XML"; hreflang=en-US; media=screen`}},
			[]LinkElem{
				{
					Anchor:   U("https://example.com/"),
					Rel:      "privacy-policy",
					Target:   U("https://example.com/privacy"),
					Title:    "Privacy",
					Type:     "application/xhtml+xml",
					HrefLang: []string{"en-us"},
					Media:    "screen",
				},
			},
		},
		{
			http.Header{"Link": {"</privacy/>; rel=privacy-policy; hreflang=en; hreflang=se"}},
			[]LinkElem{{
				Rel:      "privacy-policy",
				Target:   U("http://x.test/privacy/"),
				HrefLang: []string{"en", "se"},
			}},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Link": {"???;rel=self"}},
			[]LinkElem{},
		},
		{
			http.Header{"Link": {"<a; rel=self"}},
			[]LinkElem{},
		},
		{
			http.Header{"Link": {"a>; rel=self"}},
			[]LinkElem{},
		},
		{
			http.Header{"Link": {"<b>;rel="}},
			[]LinkElem{},
		},
		{
			http.Header{"Link": {`<b>;rel="`}},
			[]LinkElem{},
		},
		{
			http.Header{"Link": {`<b>;rel=""`}},
			[]LinkElem{},
		},
		{
			http.Header{"Link": {"<b>;=;rel=next"}},
			[]LinkElem{},
		},
		{
			http.Header{"Link": {"<b>;;;rel=next"}},
			[]LinkElem{{Rel: "next", Target: U("http://x.test/b")}},
		},
		{
			http.Header{"Link": {"<b qux>; rel=next"}},
			[]LinkElem{{Rel: "next", Target: U("http://x.test/b%20qux")}},
		},
		{
			http.Header{"Link": {"<b>; rel=next; type=text/html"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Type:   "text/html",
				},
			},
		},
		{
			http.Header{"Link": {"<b>; rel=next prefetch; hreflang=en"}},
			[]LinkElem{
				{
					Rel:      "next",
					Target:   U("http://x.test/b"),
					HrefLang: []string{"en"},
					Ext:      map[string]string{"prefetch": ""},
				},
			},
		},
		{
			http.Header{"Link": {"<b>; rel = next; title = Hello; title* = Goodbye"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Title:  "Hello",
				},
			},
		},
		{
			http.Header{"Link": {`<b>; rel = next; title* = Hello; title = "Goodbye"`}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Title:  "Goodbye",
				},
			},
		},
		{
			http.Header{"Link": {"<b>; rel = next; myAttr*"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Ext:    map[string]string{},
				},
			},
		},
		{
			http.Header{"Link": {"<b>; rel = next; myAttr* = ''"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Ext:    map[string]string{},
				},
			},
		},
		{
			http.Header{"Link": {"<b>; rel = next; myAttr = myValue; myAttr* = ''"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Ext:    map[string]string{"myattr": "myValue"},
				},
			},
		},
		{
			http.Header{"Link": {"<b>; rel = next; myAttr* = %%%%%%; myAttr = myValue"}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Ext:    map[string]string{"myattr": "myValue"},
				},
			},
		},
		{
			http.Header{"Link": {"<://example.com/>; rel=up"}},
			[]LinkElem{},
		},
		{
			http.Header{"Link": {`<b>; anchor="://example.com/"; rel=up`}},
			[]LinkElem{},
		},
		{
			// "occurrences after the first MUST be ignored by parsers"
			http.Header{"Link": {"<b>; rel=up; title=Hello; title=Goodbye"}},
			[]LinkElem{
				{Rel: "up", Target: U("http://x.test/b"), Title: "Hello"},
			},
		},
		{
			http.Header{"Link": {"<b>; rel=up; title*=Hello"}},
			[]LinkElem{
				{Rel: "up", Target: U("http://x.test/b")},
			},
		},
		{
			http.Header{"Link": {`<b>; rel=" next prefetch"; hreflang=en; extra`}},
			[]LinkElem{
				{
					Rel:      "next",
					Target:   U("http://x.test/b"),
					HrefLang: []string{"en"},
					Ext:      map[string]string{"extra": ""},
				},
				{
					Rel:      "prefetch",
					Target:   U("http://x.test/b"),
					HrefLang: []string{"en"},
					Ext:      map[string]string{"extra": ""},
				},
			},
		},
		{
			// RFC 8288 requires us to ignore these duplicates.
			http.Header{"Link": {
				`<b>;rel=next;media=screen;title=B;title*=UTF-8''BB;type=text/html; rel=prev;media=print;title=C;title*=UTF-8''CC;type=text/xml`,
			}},
			[]LinkElem{
				{
					Rel:    "next",
					Target: U("http://x.test/b"),
					Media:  "screen",
					Title:  "BB",
					Type:   "text/html",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			// When debugging failures of this test, it may be useful to
			// temporarily replace %#v with %+v in checkParse, to see the URLs.
			checkParse(t, test.header, test.result, Link(test.header, U(testBase)))
		})
	}
}

func TestSetLink(t *testing.T) {
	tests := []struct {
		input  []LinkElem
		result http.Header
	}{
		{
			nil,
			http.Header{},
		},
		{
			// Link permits an empty list.
			[]LinkElem{},
			http.Header{"Link": {""}},
		},
		{
			[]LinkElem{{Rel: "next", Target: U("baz")}},
			http.Header{"Link": {"<baz>; rel=next"}},
		},
		{
			[]LinkElem{
				{
					Anchor:   U("https://example.com/"),
					Rel:      "privacy-policy",
					Target:   U("https://example.com/privacy"),
					Title:    "Privacy",
					Type:     "application/xhtml+xml",
					HrefLang: []string{"en-US", "en-GB"},
					Ext:      map[string]string{"foo": "bar"},
				},
			},
			http.Header{"Link": {`<https://example.com/privacy>; anchor="https://example.com/"; rel=privacy-policy; title="Privacy"; type="application/xhtml+xml"; hreflang=en-US; hreflang=en-GB; foo=bar`}},
		},
		{
			[]LinkElem{
				{
					Rel:    "privacy-policy",
					Target: U("https://example.com/privacy"),
					Ext: map[string]string{
						"anchor":   "azaza",
						"rel":      "azaza",
						"Rel":      "azaza",
						"title":    "azaza",
						"title*":   "azaza",
						"type":     "azaza",
						"hreflang": "azaza",
						"media":    "azaza",
						"Media":    "azaza",
					},
				},
			},
			http.Header{"Link": {`<https://example.com/privacy>; rel=privacy-policy`}},
		},
		{
			[]LinkElem{
				{
					Rel:    "foo",
					Target: U(""),
					Title:  "Hello",
					Ext:    map[string]string{"description": "Hello"},
				},
			},
			http.Header{"Link": {`<>; rel=foo; title="Hello"; description=Hello`}},
		},
		{
			[]LinkElem{
				{
					Rel:    "foo",
					Target: U(""),
					Title:  "Hello World",
					Ext:    map[string]string{"description": "Hello World"},
				},
			},
			http.Header{"Link": {`<>; rel=foo; title="Hello World"; description="Hello World"`}},
		},
		{
			[]LinkElem{
				{
					Rel:    "foo",
					Target: U(""),
					Title:  "Hello, World",
					Ext:    map[string]string{"description": "Hello, World"},
				},
			},
			http.Header{"Link": {`<>; rel=foo; title*=UTF-8''Hello%2C%20World; title="Hello, World"; description*=UTF-8''Hello%2C%20World; description="Hello, World"`}},
		},
		{
			[]LinkElem{
				{
					Rel:    "foo",
					Target: U(""),
					Title:  "Hellõ",
					Ext:    map[string]string{"description": "Hellõ"},
				},
			},
			http.Header{"Link": {`<>; rel=foo; title*=UTF-8''Hell%C3%B5; description*=UTF-8''Hell%C3%B5`}},
		},
		{
			[]LinkElem{
				{
					Rel:    "foo",
					Target: U(""),
					Ext:    map[string]string{"description*": "Hellõ"},
				},
			},
			http.Header{"Link": {`<>; rel=foo; description*=UTF-8''Hell%C3%B5`}},
		},
		{
			[]LinkElem{{Rel: "next prefetch", Target: U("/products/124")}},
			http.Header{"Link": {`</products/124>; rel="next prefetch"`}},
		},
		{
			[]LinkElem{{Target: U("a")}, {Target: U("b")}, {Target: U("c")}},
			http.Header{"Link": {`<a>; rel="", <b>; rel="", <c>; rel=""`}},
		},
		{
			[]LinkElem{
				{Rel: "up", Target: U("a"), Ext: map[string]string{"qux": ""}},
			},
			http.Header{"Link": {`<a>; rel=up; qux=""`}},
		},
		{
			[]LinkElem{
				{Rel: "up", Target: U("a"), Ext: map[string]string{"Qux*": ""}},
			},
			http.Header{"Link": {`<a>; rel=up; Qux=""`}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetLink(header, test.input)
			checkGenerate(t, test.input, test.result, header)
		})
	}
}

func TestLinkFuzz(t *testing.T) {
	checkFuzz(t, "Link", baseLink, SetLink)
}

func TestLinkRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetLink, baseLink,
		[]LinkElem{{
			Anchor:   &url.URL{},
			Rel:      "lower token | lower URL",
			Target:   &url.URL{},
			Title:    "token | quotable | UTF-8 | empty",
			Type:     "lower token/token | empty",
			HrefLang: []string{"lower token"},
			Media:    "token | empty",
			Ext: map[string]string{
				"lower token without *": "token | quotable | UTF-8 | empty",
			},
		}},
	)
}

const (
	linkSimple  = `</chapter/4>; rel=next`
	linkComplex = `</chapter/4>; rel="next prefetch", </chapter/2>; rel=prev, </chapter/preface>; rel=start; title="Preface to the Second Edition of the \"Grand Book of Protocols\"", <../>; rel=up, <https://example.com/help>; rel=help; title*=UTF-8'en'Reader%20help, </dark.css>; rel="alternate stylesheet"; type="text/css"; media=screen`
)

func BenchmarkLinkSimple(b *testing.B) {
	base := U(testBase)
	header := http.Header{"Link": {linkSimple}}
	for i := 0; i < b.N; i++ {
		Link(header, base)
	}
}

func BenchmarkLinkComplex(b *testing.B) {
	base := U(testBase)
	header := http.Header{"Link": {linkComplex}}
	for i := 0; i < b.N; i++ {
		Link(header, base)
	}
}

func BenchmarkSetLinkSimple(b *testing.B) {
	base := U(testBase)
	header := http.Header{"Link": {linkSimple}}
	parsed := Link(header, base)
	for i := 0; i < b.N; i++ {
		SetLink(header, parsed)
	}
}

func BenchmarkSetLinkComplex(b *testing.B) {
	base := U(testBase)
	header := http.Header{"Link": {linkComplex}}
	parsed := Link(header, base)
	for i := 0; i < b.N; i++ {
		SetLink(header, parsed)
	}
}

// Adapt Link to the interface expected by checkFuzz and checkRoundTrip.
func baseLink(h http.Header) []LinkElem {
	return Link(h, U(testBase))
}

func U(s string) *url.URL {
	parsed, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return parsed
}
