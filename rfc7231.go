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

// Vary parses the Vary header from h (RFC 7231 Section 7.1.4), returning a map
// where keys are header names, canonicalized with http.CanonicalHeaderKey,
// and values are all true. A wildcard (Vary: *) is returned as map[*:true],
// so it must be checked explicitly.
func Vary(h http.Header) map[string]bool {
	var names map[string]bool
	for v, vs := iterElems("", h["Vary"]); vs != nil; v, vs = iterElems(v, vs) {
		var name string
		name, v = consumeItem(v, 0)
		name = http.CanonicalHeaderKey(name)
		if names == nil {
			names = make(map[string]bool)
		}
		names[name] = true
	}
	return names
}

// SetVary replaces the Vary header in h.
// Each key in names must be a valid field-name as per RFC 7230 Section 3.2.
// Names mapping to false are ignored. See also AddVary.
func SetVary(h http.Header, names map[string]bool) {
	h.Set("Vary", buildVary(names))
}

// AddVary is like SetVary but appends instead of replacing.
func AddVary(h http.Header, names map[string]bool) {
	h.Add("Vary", buildVary(names))
}

func buildVary(names map[string]bool) string {
	b := &strings.Builder{}
	for name, value := range names {
		if !value {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(name)
	}
	return b.String()
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

// RetryAfter parses the Retry-After header from h (RFC 7231 Section 7.1.3).
// When it is specified as delay seconds, those are added to the Date header
// if one exists in h, otherwise to the current time. If the header cannot
// be parsed, a zero Time is returned.
func RetryAfter(h http.Header) time.Time {
	v := h.Get("Retry-After")
	if v == "" {
		return time.Time{}
	}
	switch v[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // delay-seconds
		seconds, err := strconv.Atoi(v)
		if err != nil {
			return time.Time{}
		}
		// Strictly speaking, RFC 7231 says "number of seconds to delay
		// after the response is received", not after it was originated (Date),
		// but the response may have been stored or processed for a long time
		// before being fed to us, so Date might even be closer than Now().
		date, err := http.ParseTime(h.Get("Date"))
		if err != nil {
			date = time.Now()
		}
		return date.Add(time.Duration(seconds) * time.Second)

	default: // HTTP-date
		date, err := http.ParseTime(v)
		if err != nil {
			return time.Time{}
		}
		return date
	}
}

// SetRetryAfter replaces the Retry-After header in h.
func SetRetryAfter(h http.Header, after time.Time) {
	h.Set("Retry-After", after.Format(http.TimeFormat))
}

// ContentType parses the Content-Type header from h (RFC 7231 Section 3.1.1.5).
// Media type and parameter names (but not values) are canonicalized to lowercase.
func ContentType(h http.Header) Par {
	ctype, _ := consumeParameterized(h.Get("Content-Type"), true)
	return ctype
}

// SetContentType replaces the Content-Type header in h.
// Media type and parameter names must be valid as per RFC 7231 Section 3.1.1.1;
// parameter values may contain any bytes except control characters.
func SetContentType(h http.Header, ctype Par) {
	b := &strings.Builder{}
	writeParameterized(b, ctype)
	h.Set("Content-Type", b.String())
}

// An AcceptElem represents one element of the Accept header
// (RFC 7231 Section 5.3.2).
type AcceptElem struct {
	Type string  // media range, canonicalized to lowercase
	Q    float32 // quality value
	// All map keys are canonicalized to lowercase.
	Params map[string]string // media type parameters
	Ext    map[string]string // extension parameters
}

// Accept parses the Accept header from h (RFC 7231 Section 5.3.2).
func Accept(h http.Header) []AcceptElem {
	var elems []AcceptElem
	for v, vs := iterElems("", h["Accept"]); vs != nil; v, vs = iterElems(v, vs) {
		elem := AcceptElem{Q: 1}
		elem.Type, v = consumeItem(v, 0)
		elem.Type = strings.ToLower(elem.Type)
		afterQ := false
		for {
			v = skipWS(v)
			if peek(v) != ';' {
				break
			}
			v = skipWS(v[1:])
			if c := peek(v); c == ';' || c == ',' || c == 0 {
				// This is an empty parameter.
				continue
			}
			var name, value string
			name, value, v = consumeParam(v, true)
			switch {
			case name == "q":
				qvalue, _ := strconv.ParseFloat(value, 32)
				elem.Q = float32(qvalue)
				afterQ = true
			case afterQ:
				if elem.Ext == nil {
					elem.Ext = make(map[string]string)
				}
				elem.Ext[name] = value
			default:
				if elem.Params == nil {
					elem.Params = make(map[string]string)
				}
				elem.Params[name] = value
			}
		}
		elems = append(elems, elem)
	}
	return elems
}

// SetAccept replaces the Accept header in h.
// In each of elems, Type must be a valid media range (RFC 7231 Section 5.3.2),
// Q must be a valid qvalue (RFC 7231 Section 5.3.1), and all map keys must be
// valid tokens (RFC 7230 Section 3.2.6); map values may contain any bytes except
// control characters.
//
// Note: Q must be set explicitly to avoid sending "q=0", meaning "not acceptable".
func SetAccept(h http.Header, elems []AcceptElem) {
	b := &strings.Builder{}
	for i, elem := range elems {
		if i > 0 {
			b.WriteString(", ")
		}
		writeParameterized(b, Par{elem.Type, elem.Params})
		if elem.Q != 1 || len(elem.Ext) > 0 {
			b.WriteString(";q=")
			// "A sender of qvalue MUST NOT generate more than three digits
			// after the decimal point."
			b.WriteString(strconv.FormatFloat(float64(elem.Q), 'g', 3, 32))
		}
		writeNullableParams(b, elem.Ext)
	}
	h.Set("Accept", b.String())
}
