package httpheader

import (
	"net/http"
	"strings"
)

// ContentDisposition parses the Content-Disposition header from h (RFC 6266),
// returning the disposition type, canonicalized to lowercase,
// and a map of disposition parameters, where keys are also lowercased.
// Parameters whose names end in an asterisk, such as 'filename*',
// are automatically decoded from RFC 5987 encoding, but their UTF-8 is not
// validated.
func ContentDisposition(h http.Header) (dtype string, params map[string]string) {
	dtype, params, _ = consumeParameterized(h.Get("Content-Disposition"))
	for name, value := range params {
		if strings.HasSuffix(name, "*") {
			decoded, err := decodeExtValue(value)
			if err != nil {
				continue
			}
			params[name] = decoded
		}
	}
	return
}

// SetContentDisposition replaces the Content-Disposition header in h.
// Parameters whose names end in an asterisk, such as 'filename*',
// are automatically encoded into RFC 5987 encoding; their value
// must already be valid UTF-8.
func SetContentDisposition(h http.Header, dtype string, params map[string]string) {
	b := &strings.Builder{}
	b.WriteString(dtype)
	// RFC 6266 Appendix D: "When a 'filename' parameter is included as a fallback
	// [...], 'filename' should occur first, due to parsing problems in some
	// existing implementations."
	if filename, ok := params["filename"]; ok {
		b.WriteString("; filename=")
		writeTokenOrQuoted(b, filename)
	}
	for name, value := range params {
		if name == "filename" { // handled above
			continue
		}
		b.WriteString("; ")
		b.WriteString(name)
		b.WriteString("=")
		if strings.HasSuffix(name, "*") {
			writeExtValue(b, value)
		} else {
			writeTokenOrQuoted(b, value)
		}
	}
	h.Set("Content-Disposition", b.String())
}
