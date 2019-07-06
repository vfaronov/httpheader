package httpheader

import (
	"net/http"
	"strings"
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
// Each of methods must be valid as per RFC 7230 Section 7.1.1.
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

// SetVary replaces the Vary header in h.
// Each of names must be a valid field-name as per RFC 7230 Section 3.2.
// See also AddVary.
func SetVary(h http.Header, names []string) {
	h.Set("Vary", strings.Join(names, ", "))
}

// AddVary is like SetVary but appends instead of replacing.
func AddVary(h http.Header, names []string) {
	h.Add("Vary", strings.Join(names, ", "))
}
