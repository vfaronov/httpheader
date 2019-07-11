package httpheader

import (
	"net/http"
	"strings"
)

// A ViaElem represents one element of the Via header (RFC 7230 Section 5.7.1).
type ViaElem struct {
	ReceivedProto string // canonicalized to include name: "HTTP/1.1", not "1.1"
	ReceivedBy    string
	Comment       string
}

// Via parses the Via header from h (RFC 7230 Section 5.7.1).
//
// BUG(vfaronov): Incorrectly parses some extravagant values of uri-host
// that do not occur in practice but are theoretically admitted by RFC 3986.
func Via(h http.Header) []ViaElem {
	var elems []ViaElem
	for v, vs := iterElems("", h["Via"]); vs != nil; v, vs = iterElems(v, vs) {
		var elem ViaElem
		elem.ReceivedProto, v = consumeItem(v)
		if strings.IndexByte(elem.ReceivedProto, '/') == -1 {
			elem.ReceivedProto = "HTTP/" + elem.ReceivedProto
		}
		v = skipWS(v)
		elem.ReceivedBy, v = consumeAgent(v)
		v = skipWS(v)
		if peek(v) == '(' {
			elem.Comment, v, _ = consumeComment(v, false)
		}
		elems = append(elems, elem)
	}
	return elems
}

// SetVia replaces the Via header in h. See also AddVia.
func SetVia(h http.Header, elems []ViaElem) {
	h.Set("Via", buildVia(elems))
}

// AddVia is like SetVia but appends instead of replacing.
func AddVia(h http.Header, elem ViaElem) {
	h.Add("Via", buildVia([]ViaElem{elem}))
}

func buildVia(elems []ViaElem) string {
	b := &strings.Builder{}
	for i, elem := range elems {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strings.TrimPrefix(elem.ReceivedProto, "HTTP/"))
		b.WriteString(" ")
		b.WriteString(elem.ReceivedBy)
		if elem.Comment != "" {
			b.WriteString(" ")
			writeComment(b, elem.Comment)
		}
	}
	return b.String()
}
