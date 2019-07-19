package httpheader

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func ExampleContentDisposition() {
	header := http.Header{"Content-Disposition": {
		`Attachment; filename="EURO rates"; filename*=utf-8''%e2%82%ac%20rates`,
	}}
	dtype, filename, _ := ContentDisposition(header)
	fmt.Println(dtype)
	fmt.Println(filename)
	// Output: attachment
	// € rates
}

func TestContentDisposition(t *testing.T) {
	tests := []struct {
		header   http.Header
		dtype    string
		filename string
		params   map[string]string
	}{
		// Valid headers.
		{
			http.Header{"Content-Disposition": {"inline"}},
			"inline", "", nil,
		},
		{
			http.Header{"Content-Disposition": {
				"attachment; Filename=hello.txt",
			}},
			"attachment", "hello.txt", nil,
		},
		{
			http.Header{"Content-Disposition": {
				`attachment; filename="hello world.txt"`,
			}},
			"attachment", "hello world.txt", nil,
		},
		{
			// This is valid because RFC 6266 is based on the old grammar of HTTP
			// with implied *LWS as per RFC 2616 Section 2.1.
			http.Header{"Content-Disposition": {
				`attachment;  filename = "hello world.txt"`,
			}},
			"attachment", "hello world.txt", nil,
		},
		{
			http.Header{"Content-Disposition": {
				"attachment; filename*=utf-8''hello.txt",
			}},
			"attachment", "hello.txt", nil,
		},
		{
			http.Header{"Content-Disposition": {
				`Attachment; Filename="Privet, mir!.txt"; Filename*=UTF-8'ru'%D0%9F%D1%80%D0%B8%D0%B2%D0%B5%D1%82%2C%20%D0%BC%D0%B8%D1%80!.txt`,
			}},
			"attachment", "Привет, мир!.txt", nil,
		},
		{
			// RFC 8187 no longer requires us to support ISO-8859-1.
			http.Header{"Content-Disposition": {
				`Attachment; Filename="Hola mundo!.txt"; Filename*=iso-8859-1'es'%a1Hola%20mundo!.txt`,
			}},
			"attachment", "Hola mundo!.txt", nil,
		},
		{
			http.Header{"Content-Disposition": {
				"attachment;\tfilename*=utf-8''hello.txt\t;filename=hello.txt",
			}},
			"attachment", "hello.txt", nil,
		},
		{
			http.Header{"Content-Disposition": {"foo;bar*=UTF-8''"}},
			"foo", "", map[string]string{"bar": ""},
		},
		{
			http.Header{"Content-Disposition": {"foo;bar*=UTF-8''baz+qux"}},
			"foo", "", map[string]string{"bar": "baz+qux"},
		},
		{
			http.Header{"Content-Disposition": {"foo;bar=baz;bar*=UTF-8''b%c3%a1z"}},
			"foo", "", map[string]string{"bar": "báz"},
		},
		{
			http.Header{"Content-Disposition": {"foo;bar*=UTF-8''b%c3%a1z;bar=baz"}},
			"foo", "", map[string]string{"bar": "báz"},
		},
		{
			// Strange though it may look, this is a valid ext-value, and it
			// overrides the plain filename.
			http.Header{"Content-Disposition": {"attachment; filename=hello.txt; filename*=UTF-8''"}},
			"attachment", "", nil,
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Content-Disposition": {"attachment; filename*="}},
			"attachment", "", nil,
		},
		{
			http.Header{"Content-Disposition": {"attachment; filename*='"}},
			"attachment", "", nil,
		},
		{
			http.Header{"Content-Disposition": {"attachment; filename*=''"}},
			"attachment", "", nil,
		},
		{
			// Bad UTF-8 after percent-decoding.
			http.Header{"Content-Disposition": {
				"attachment; filename*=utf-8''%81%82%83%84.txt",
			}},
			"attachment", "\x81\x82\x83\x84.txt", nil,
		},
		{
			// Bad percent encoding.
			http.Header{"Content-Disposition": {
				"attachment; filename*=utf-8''%%%%%%%%.txt",
			}},
			"attachment", "", nil,
		},
		{
			// Only two apostrophe-separated fields instead of three.
			http.Header{"Content-Disposition": {
				"attachment; filename*=utf-8'hello.txt",
			}},
			"attachment", "", nil,
		},
		{
			// Not RFC 8187 syntax at all.
			http.Header{"Content-Disposition": {
				"attachment; filename*=%81%82%83%84.txt",
			}},
			"attachment", "", nil,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			dtype, filename, params := ContentDisposition(test.header)
			checkParse(t, test.header,
				test.dtype, dtype,
				test.filename, filename,
				test.params, params,
			)
		})
	}
}

func ExampleSetContentDisposition() {
	header := http.Header{}
	SetContentDisposition(header, "attachment", "Résumé.docx", nil)
	header.Write(os.Stdout)
	// Output: Content-Disposition: attachment; filename*=UTF-8''R%C3%A9sum%C3%A9.docx
}

func TestSetContentDisposition(t *testing.T) {
	tests := []struct {
		dtype    string
		filename string
		params   map[string]string
		result   http.Header
	}{
		{
			"attachment", "foo; bar", nil,
			http.Header{"Content-Disposition": {
				`attachment; filename="foo; bar"; filename*=UTF-8''foo%3B%20bar`,
			}},
		},
		{
			"attachment", "báz.txt", map[string]string{"filename": "baz.txt"},
			http.Header{"Content-Disposition": {
				"attachment; filename*=UTF-8''b%C3%A1z.txt",
			}},
		},
		{
			"inline", "", map[string]string{"foo": "bar baz"},
			http.Header{"Content-Disposition": {`inline; foo="bar baz"`}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			input := []interface{}{test.dtype, test.filename, test.params}
			header := http.Header{}
			SetContentDisposition(header, test.dtype, test.filename, test.params)
			checkGenerate(t, input, test.result, header)
		})
	}
}

func TestContentDispositionRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetContentDisposition, ContentDisposition,
		"lower token",
		"quotable | UTF-8 | empty",
		map[string]string{"lower token without *": "quotable | UTF-8 | empty"},
	)
}

func BenchmarkContentDispositionSimple(b *testing.B) {
	header := http.Header{"Content-Disposition": {`attachment; filename="Privet mir.txt"; filename*=UTF-8''%D0%9F%D1%80%D0%B8%D0%B2%D0%B5%D1%82%20%D0%BC%D0%B8%D1%80.txt`}}
	for i := 0; i < b.N; i++ {
		ContentDisposition(header)
	}
}

func BenchmarkContentDispositionComplex(b *testing.B) {
	header := http.Header{"Content-Disposition": {`attachment; filename="Privet mir.txt"; filename*=UTF-8''%D0%9F%D1%80%D0%B8%D0%B2%D0%B5%D1%82%20%D0%BC%D0%B8%D1%80.txt; type="text/html"; description*=UTF-8''%D0%9F%D1%80%D0%B8%D0%B2%D0%B5%D1%82%D1%81%D1%82%D0%B2%D0%B8%D0%B5`}}
	for i := 0; i < b.N; i++ {
		ContentDisposition(header)
	}
}
