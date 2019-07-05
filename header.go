package httpheader

import (
	"net/http"
	"strings"
)

// Allow returns a slice of method names from the Allow field in h
// (RFC 7231 Section 7.4.1).
//
// If there is no such field in h, Allow returns nil.
// If the field is present but empty (meaning all methods are disallowed),
// Allow returns a non-nil slice of length 0.
func Allow(h http.Header) []string {
	var methods []string
	for v, vs := toNextElem("", h["Allow"]); vs != nil; v, vs = toNextElem(v, vs) {
		var tok string
		if tok, v = token(v); tok != "" {
			methods = append(methods, tok)
		}
	}
	if methods == nil && len(h["Allow"]) > 0 {
		methods = make([]string, 0)
	}
	return methods
}

// SetAllow sets the Allow field (RFC 7231 Section 7.4.1) in h.
func SetAllow(h http.Header, methods []string) {
	h.Set("Allow", strings.Join(methods, ", "))
}

// Vary returns a slice of names from the Vary field in h
// (RFC 7231 Section 7.1.4).
// Names are canonicalized with http.CanonicalHeaderKey.
// A wildcard (Vary: *) is returned as a slice of 1 element.
func Vary(h http.Header) []string {
	var fields []string
	for v, vs := toNextElem("", h["Vary"]); vs != nil; v, vs = toNextElem(v, vs) {
		var tok string
		if tok, v = token(v); tok != "" {
			fields = append(fields, http.CanonicalHeaderKey(tok))
		}
	}
	return fields
}

// SetVary sets the Vary field in h (RFC 7231 Section 7.1.4). See also AddVary.
func SetVary(h http.Header, fields []string) {
	h.Set("Vary", strings.Join(fields, ", "))
}

// AddVary appends to the Vary field in h (RFC 7231 Section 7.1.4).
func AddVary(h http.Header, fields []string) {
	h.Add("Vary", strings.Join(fields, ", "))
}

// Via returns a slice of entries from the Via field in h (RFC 7230 Section 5.7.1).
func Via(h http.Header) []ViaEntry {
	var entries []ViaEntry
	for v, vs := toNextElem("", h["Via"]); vs != nil; v, vs = toNextElem(v, vs) {
		var entry ViaEntry
		entry.ReceivedProto, v = chomp(v)
		if entry.ReceivedProto == "" {
			continue
		}
		if !strings.ContainsRune(entry.ReceivedProto, '/') {
			entry.ReceivedProto = "HTTP/" + entry.ReceivedProto
		}
		entry.ReceivedBy, v = chomp(v)
		if entry.ReceivedBy == "" {
			continue
		}
		if peek(v) == '(' {
			entry.Comment, v = comment(v)
		}
		entries = append(entries, entry)
	}
	return entries
}

// SetVia sets the Via field in h (RFC 7230 Section 5.7.1). See also AddVia.
func SetVia(h http.Header, entries []ViaEntry) {
	h.Set("Via", makeVia(entries))
}

// AddVia appends to the Via field in h (RFC 7230 Section 5.7.1).
func AddVia(h http.Header, entries []ViaEntry) {
	h.Add("Via", makeVia(entries))
}

func makeVia(entries []ViaEntry) string {
	b := &strings.Builder{}
	for i, entry := range entries {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strings.TrimPrefix(entry.ReceivedProto, "HTTP/"))
		b.WriteString(" ")
		b.WriteString(entry.ReceivedBy)
		if entry.Comment != "" {
			b.WriteString(" ")
			writeDelimited(b, entry.Comment, '(', ')')
		}
	}
	return b.String()
}

// A ViaEntry represents one element of the Via field (RFC 7230 Section 5.7.1).
type ViaEntry struct {
	ReceivedProto string // always includes name: "HTTP/1.1", not "1.1"
	ReceivedBy    string
	Comment       string
}
