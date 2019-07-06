package httpheader

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Allow parses the Allow header from h (RFC 7231 Section 7.4.1).
//
// If there is no such header in h, Allow returns nil.
// If the header is present but empty (meaning all methods are disallowed),
// Allow returns a non-nil slice of length 0.
func Allow(h http.Header) []string {
	var methods []string
	for v, vs := iterElems("", h["Allow"]); vs != nil; v, vs = iterElems(v, vs) {
		var method string
		method, v = consumeItem(v)
		if method != "" {
			methods = append(methods, method)
		}
	}
	if methods == nil && h["Allow"] != nil {
		methods = make([]string, 0)
	}
	return methods
}

// SetAllow replaces the Allow header in h.
func SetAllow(h http.Header, methods []string) {
	h.Set("Allow", strings.Join(methods, ", "))
}

// Vary parses the Vary header from h (RFC 7231 Section 7.1.4).
// Parsed names are canonicalized with http.CanonicalHeaderKey.
// A wildcard (Vary: *) is returned as a slice of 1 element.
func Vary(h http.Header) []string {
	var names []string
	for v, vs := iterElems("", h["Vary"]); vs != nil; v, vs = iterElems(v, vs) {
		var name string
		name, v = consumeItem(v)
		if name != "" {
			names = append(names, http.CanonicalHeaderKey(name))
		}
	}
	return names
}

// SetVary replaces the Vary header in h. See also AddVary.
func SetVary(h http.Header, names []string) {
	h.Set("Vary", strings.Join(names, ", "))
}

// AddVary appends to the Vary header in h.
func AddVary(h http.Header, names []string) {
	h.Add("Vary", strings.Join(names, ", "))
}

// Via parses the Via header from h (RFC 7230 Section 5.7.1).
func Via(h http.Header) []ViaElem {
	var elems []ViaElem
	for v, vs := iterElems("", h["Via"]); vs != nil; v, vs = iterElems(v, vs) {
		var elem ViaElem
		elem.ReceivedProto, v = consumeItem(v)
		if strings.IndexByte(elem.ReceivedProto, '/') == -1 {
			elem.ReceivedProto = "HTTP/" + elem.ReceivedProto
		}
		v = skipWS(v)
		elem.ReceivedBy, v = consumeItem(v)
		v = skipWS(v)
		if peek(v) == '(' {
			elem.Comment, v = consumeComment(v)
		}
		elems = append(elems, elem)
	}
	return elems
}

// SetVia replaces the Via header in h. See also AddVia.
func SetVia(h http.Header, elems []ViaElem) {
	h.Set("Via", buildVia(elems))
}

// AddVia appends to the Via header in h.
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
			writeDelimited(b, elem.Comment, '(', ')')
		}
	}
	return b.String()
}

// A ViaElem represents one element of the Via header (RFC 7230 Section 5.7.1).
type ViaElem struct {
	ReceivedProto string // always includes name: "HTTP/1.1", not "1.1"
	ReceivedBy    string
	Comment       string
}

// Warning parses the Warning header from h (RFC 7234 Section 5.5).
func Warning(h http.Header) []WarningElem {
	var elems []WarningElem
	for v, vs := iterElems("", h["Warning"]); vs != nil; v, vs = iterElems(v, vs) {
		var elem WarningElem
		var codeStr string
		codeStr, v = consumeItem(v)
		elem.Code, _ = strconv.Atoi(codeStr)
		v = skipWS(v)
		elem.Agent, v = consumeItem(v)
		v = skipWS(v)
		elem.Text, v = consumeQuoted(v)
		v = skipWS(v)
		if peek(v) == '"' {
			nextQuote := strings.IndexByte(v[1:], '"')
			if nextQuote > 0 {
				elem.Date, _ = http.ParseTime(v[1:nextQuote+1])
				v = v[nextQuote+1:]
			}
		}
		elems = append(elems, elem)
	}
	return elems
}

// A WarningElem represents one element of the Warning header
// (RFC 7234 Section 5.5).
type WarningElem struct {
	Code  int
	Agent string
	Text  string
	Date  time.Time // zero if missing
}

// SetWarning replaces the Warning header in h. See also AddWarning.
func SetWarning(h http.Header, elems []WarningElem) {
	h.Set("Warning", buildWarning(elems))
}

// AddWarning appends to the Warning header in h.
func AddWarning(h http.Header, elem WarningElem) {
	h.Add("Warning", buildWarning([]WarningElem{elem}))
}

func buildWarning(elems []WarningElem) string {
	b := &strings.Builder{}
	for i, elem := range elems {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(elem.Code))
		b.WriteString(" ")
		b.WriteString(elem.Agent)
		b.WriteString(" ")
		writeDelimited(b, elem.Text, '"', '"')
		if !elem.Date.IsZero() {
			b.WriteString(` "`)
			b.WriteString(elem.Date.Format(http.TimeFormat))
			b.WriteString(`"`)
		}
	}
	return b.String()
}
