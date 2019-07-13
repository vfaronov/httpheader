package httpheader

import (
	"fmt"
	"net/http"
	"testing"
)

func ExampleContentDisposition() {
	header := http.Header{"Content-Disposition": {
		`Attachment; filename="EURO rates"; filename*=utf-8''%e2%82%ac%20rates`,
	}}
	dtype, params := ContentDisposition(header)
	fmt.Println(dtype)
	fmt.Println(params["filename*"])
	// Output: attachment
	// € rates
}

func TestContentDisposition(t *testing.T) {
	tests := []struct {
		header http.Header
		dtype  string
		params map[string]string
	}{
		// Valid headers.
		{
			http.Header{"Content-Disposition": {"inline"}},
			"inline",
			nil,
		},
		{
			http.Header{"Content-Disposition": {
				"attachment; Filename=hello.txt",
			}},
			"attachment",
			map[string]string{"filename": "hello.txt"},
		},
		{
			http.Header{"Content-Disposition": {
				`attachment; filename="hello world.txt"`,
			}},
			"attachment",
			map[string]string{"filename": "hello world.txt"},
		},
		{
			// This is valid because RFC 6266 is based on the old grammar of HTTP
			// with implied *LWS as per RFC 2616 Section 2.1.
			http.Header{"Content-Disposition": {
				`attachment;  filename = "hello world.txt"`,
			}},
			"attachment",
			map[string]string{"filename": "hello world.txt"},
		},
		{
			http.Header{"Content-Disposition": {
				"attachment; filename*=utf-8''hello.txt",
			}},
			"attachment",
			map[string]string{"filename*": "hello.txt"},
		},
		{
			http.Header{"Content-Disposition": {
				`Attachment; Filename="Privet, mir!.txt"; Filename*=UTF-8'ru'%D0%9F%D1%80%D0%B8%D0%B2%D0%B5%D1%82%2C%20%D0%BC%D0%B8%D1%80!.txt`,
			}},
			"attachment",
			map[string]string{
				"filename":  "Privet, mir!.txt",
				"filename*": "Привет, мир!.txt",
			},
		},
		{
			// RFC 8187 no longer requires us to support ISO-8859-1.
			http.Header{"Content-Disposition": {
				`Attachment; Filename="Hola mundo!.txt"; Filename*=iso-8859-1'es'%a1Hola%20mundo!.txt`,
			}},
			"attachment",
			map[string]string{
				"filename":  "Hola mundo!.txt",
				"filename*": "iso-8859-1'es'%a1Hola%20mundo!.txt",
			},
		},
		{
			http.Header{"Content-Disposition": {
				"attachment;\tfilename*=utf-8''hello.txt\t;filename=hello.txt",
			}},
			"attachment",
			map[string]string{"filename*": "hello.txt", "filename": "hello.txt"},
		},
		{
			http.Header{"Content-Disposition": {"foo;bar*=UTF-8''"}},
			"foo",
			map[string]string{"bar*": ""},
		},
		{
			http.Header{"Content-Disposition": {"foo;bar*=UTF-8''baz+qux"}},
			"foo",
			map[string]string{"bar*": "baz+qux"},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Content-Disposition": {"attachment; filename*="}},
			"attachment",
			map[string]string{"filename*": ""},
		},
		{
			http.Header{"Content-Disposition": {"attachment; filename*='"}},
			"attachment",
			map[string]string{"filename*": "'"},
		},
		{
			http.Header{"Content-Disposition": {"attachment; filename*=''"}},
			"attachment",
			map[string]string{"filename*": "''"},
		},
		{
			// Bad UTF-8 after percent-decoding.
			http.Header{"Content-Disposition": {
				"attachment; filename*=utf-8''%81%82%83%84.txt",
			}},
			"attachment",
			map[string]string{"filename*": "\x81\x82\x83\x84.txt"},
		},
		{
			// Bad percent encoding.
			http.Header{"Content-Disposition": {
				"attachment; filename*=utf-8''%%%%%%%%.txt",
			}},
			"attachment",
			map[string]string{"filename*": "utf-8''%%%%%%%%.txt"},
		},
		{
			// Only two apostrophe-separated fields instead of three.
			http.Header{"Content-Disposition": {
				"attachment; filename*=utf-8'hello.txt",
			}},
			"attachment",
			map[string]string{"filename*": "utf-8'hello.txt"},
		},
		{
			// Not RFC 8187 syntax at all.
			http.Header{"Content-Disposition": {
				"attachment; filename*=%81%82%83%84.txt",
			}},
			"attachment",
			map[string]string{"filename*": "%81%82%83%84.txt"},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			dtype, params := ContentDisposition(test.header)
			checkParse(t, test.header, test.dtype, dtype, test.params, params)
		})
	}
}

func ExampleSetContentDisposition() {
	header := http.Header{}
	SetContentDisposition(header, "attachment", map[string]string{
		"filename*": "Résumé.docx",
		// ASCII-only fallback, as recommended by RFC 6266 Appendix D.
		"filename": "Resume.docx",
	})
	fmt.Printf("%+q", header)
	// Output: map["Content-Disposition":["attachment; filename=Resume.docx; filename*=UTF-8''R%C3%A9sum%C3%A9.docx"]]
}

func TestSetContentDisposition(t *testing.T) {
	tests := []struct {
		dtype  string
		params map[string]string
		result http.Header
	}{
		{
			"attachment",
			map[string]string{"filename*": "foo = bar"},
			http.Header{"Content-Disposition": {
				"attachment; filename*=UTF-8''foo%20%3D%20bar",
			}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			input := []interface{}{test.dtype, test.params}
			header := http.Header{}
			SetContentDisposition(header, test.dtype, test.params)
			checkGenerate(t, input, test.result, header)
		})
	}
}

func TestContentDispositionFuzz(t *testing.T) {
	checkFuzz(t, "Content-Disposition", ContentDisposition, SetContentDisposition)
}

func TestContentDispositionRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetContentDisposition, ContentDisposition,
		"lower token",
		map[string]string{
			"lower token":        "quotable | empty",
			"lower token plus *": "quotable | UTF-8 | empty",
		},
	)
}
