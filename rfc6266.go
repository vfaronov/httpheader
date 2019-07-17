package httpheader

import (
	"net/http"
	"strings"
)

// ContentDisposition parses the Content-Disposition header from h (RFC 6266),
// returning the disposition type, the value of the 'filename' parameter (if any),
// and a map of any other parameters.
//
// Any 'filename*' parameter is decoded from RFC 8187 encoding, and overrides
// 'filename'. Similarly for any other parameter whose name ends in an asterisk.
// UTF-8 is not validated in such strings.
func ContentDisposition(h http.Header) (dtype, filename string, params map[string]string) {
	v := h.Get("Content-Disposition")
	dtype, v = consumeItem(v)
	dtype = strings.ToLower(dtype)
ParamsLoop:
	for {
		var name, value string
		name, value, v = consumeParam(v)
		switch name {
		case "":
			break ParamsLoop
		case "filename":
			if filename == "" { // not set from 'filename*' yet
				filename = value
			}
		case "filename*":
			if decoded, _, err := DecodeExtValue(value); err == nil {
				filename = decoded
			}
		default:
			params = insertVariform(params, name, value)
		}
	}
	return
}

// SetContentDisposition replaces the Content-Disposition header in h.
//
// If filename is not empty, it must be valid UTF-8, which is serialized into
// a 'filename' parameter in plain ASCII, or a 'filename*' parameter in RFC 8187
// encoding, or both, depending on what characters it contains.
//
// Similarly, if params contains a 'qux' or 'qux*' key, it will be serialized into
// a 'qux' and/or 'qux*' parameter depending on its contents; the asterisk
// in the key is ignored. Any 'filename' or 'filename*' in params is skipped.
func SetContentDisposition(h http.Header, dtype, filename string, params map[string]string) {
	b := &strings.Builder{}
	write(b, dtype)
	if filename != "" {
		writeVariform(b, "filename", filename)
	}
	for name, value := range params {
		if strings.ToLower(strings.TrimSuffix(name, "*")) == "filename" {
			continue
		}
		writeVariform(b, name, value)
	}
	h.Set("Content-Disposition", b.String())
}
