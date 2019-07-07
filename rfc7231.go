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
		method, v = consumeItem(v, 0)
		methods = append(methods, method)
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
		name, v = consumeItem(v, 0)
		names = append(names, http.CanonicalHeaderKey(name))
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

// A Product contains software information as found in the User-Agent
// and Server headers (RFC 7231 Section 5.5.3 and Section 7.4.2).
// If multiple comments are associated with a product, they are concatenated
// with a "; " separator.
type Product struct {
	Name, Version, Comment string
}

// UserAgent parses the User-Agent header from h (RFC 7231 Section 5.5.3).
func UserAgent(h http.Header) []Product {
	return parseProducts(h.Get("User-Agent"))
}

// SetUserAgent replaces the User-Agent header in h.
// In each of products, Name must be a valid token (RFC 7230 Section 7.7.1)
// and Version must be a valid token if not empty; Comment may contain any bytes
// except control characters.
func SetUserAgent(h http.Header, products []Product) {
	h.Set("User-Agent", serializeProducts(products))
}

// Server parses the Server header from h (RFC 7231 Section 7.4.2).
func Server(h http.Header) []Product {
	return parseProducts(h.Get("Server"))
}

// SetServer replaces the Server header in h.
// Each of products must obey the same requirements as for SetUserAgent.
func SetServer(h http.Header, products []Product) {
	h.Set("Server", serializeProducts(products))
}

func parseProducts(v string) []Product {
	var products []Product
	for v != "" {
		var product Product
		product.Name, v = consumeItem(v, '/')
		if product.Name == "" {
			// Avoid infinite loop.
			v = v[1:]
			continue
		}
		if peek(v) == '/' {
			product.Version, v = consumeItem(v[1:], 0)
		}
		// Collect all comments for this product.
		for {
			v = skipWS(v)
			if peek(v) != '(' {
				break
			}
			var comment string
			comment, v = consumeComment(v)
			if product.Comment == "" {
				product.Comment = comment
			} else {
				product.Comment += "; " + comment
			}
		}
		products = append(products, product)
	}
	return products
}

func serializeProducts(products []Product) string {
	b := &strings.Builder{}
	for i, product := range products {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(product.Name)
		if product.Version != "" {
			b.WriteString("/")
			b.WriteString(product.Version)
		}
		if product.Comment != "" {
			b.WriteString(" ")
			writeComment(b, product.Comment)
		}
	}
	return b.String()
}
