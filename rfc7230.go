package httpheader

import (
	"net/http"
	"strings"
)

// A ViaElem represents one element of the Via header (RFC 7230 Section 5.7.1).
type ViaElem struct {
	ReceivedProto string
	ReceivedBy    string
	Comment       string
}

// Via parses the Via header from h (RFC 7230 Section 5.7.1).
//
// ReceivedProto in returned elements is canonicalized to always include name:
// ``1.1'' becomes ``HTTP/1.1''. As a special case, ``2'' and ``HTTP/2'' become
// ``HTTP/2.0''.
//
// BUG(vfaronov): Incorrectly parses some extravagant values of uri-host
// that do not occur in practice but are theoretically admitted by RFC 3986.
func Via(h http.Header) []ViaElem {
	var elems []ViaElem
	for v, vs := iterElems("", h["Via"]); v != ""; v, vs = iterElems(v, vs) {
		var elem ViaElem
		elem.ReceivedProto, v = consumeItem(v)
		elem.ReceivedProto = canonicalProto(elem.ReceivedProto)
		v = skipWS(v)
		elem.ReceivedBy, v = consumeReceivedBy(v)
		v = skipWS(v)
		if peek(v) == '(' {
			elem.Comment, v = consumeComment(v)
		}
		elems = append(elems, elem)
	}
	return elems
}

func canonicalProto(proto string) string {
	// Special-case typical values to avoid allocating them every time.
	// Also use this opportunity to canonicalize "2" to "2.0",
	// which is what's used by net/http in Request.Proto and Response.Proto
	// (see also RFC 7540 Errata 4663 and its discussion on ietf-http-wg@w3.org).
	switch proto {
	case "1.0":
		return "HTTP/1.0"
	case "1.1":
		return "HTTP/1.1"
	case "2.0", "2", "HTTP/2":
		return "HTTP/2.0"
	default:
		if strings.IndexByte(proto, '/') != -1 {
			return proto
		}
		return "HTTP/" + proto
	}
}

func consumeReceivedBy(v string) (by, newv string) {
	// This is tricky because received-by can contain commas, semicolons, equal
	// signs (see test cases) or even be empty, if you read the grammar literally
	// (reg-name may be empty). The reg-name cases are too much for me right now,
	// but it's easy to handle the IP-Literal case: it's delimited by brackets
	// and never contains brackets.
	if peek(v) == '[' {
		if end := strings.IndexByte(v, ']'); end >= 0 {
			var maybePort string
			maybePort, newv = consumeItem(v[end+1:])
			by = v[:end+1+len(maybePort)]
			return
		}
	}
	return consumeItem(v)
}

// SetVia replaces the Via header in h. See also AddVia.
func SetVia(h http.Header, elems []ViaElem) {
	if len(elems) == 0 {
		h.Del("Via")
		return
	}
	h.Set("Via", buildVia(elems))
}

// AddVia is like SetVia but appends instead of replacing.
func AddVia(h http.Header, elems ...ViaElem) {
	if len(elems) == 0 {
		return
	}
	h.Add("Via", buildVia(elems))
}

func buildVia(elems []ViaElem) string {
	b := &strings.Builder{}
	for i, elem := range elems {
		if i > 0 {
			write(b, ", ")
		}
		write(b, strings.TrimPrefix(elem.ReceivedProto, "HTTP/"),
			" ", elem.ReceivedBy)
		if elem.Comment != "" {
			write(b, " ")
			writeComment(b, elem.Comment)
		}
	}
	return b.String()
}
