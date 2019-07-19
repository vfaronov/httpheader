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
	values := h["Allow"]
	if values == nil {
		return nil
	}
	methods := make([]string, 0, estimateElems(values))
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		var method string
		method, v = consumeItem(v)
		methods = append(methods, method)
	}
	return methods
}

// SetAllow replaces the Allow header in h (RFC 7231 Section 7.4.1).
func SetAllow(h http.Header, methods []string) {
	h.Set("Allow", strings.Join(methods, ", "))
}

// Vary parses the Vary header from h (RFC 7231 Section 7.1.4), returning a map
// where keys are header names, canonicalized with http.CanonicalHeaderKey,
// and values are all true. A wildcard (Vary: *) is returned as map[*:true],
// so it must be checked explicitly.
func Vary(h http.Header) map[string]bool {
	values := h["Vary"]
	if values == nil {
		return nil
	}
	names := make(map[string]bool)
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		var name string
		name, v = consumeItem(v)
		name = http.CanonicalHeaderKey(name)
		names[name] = true
	}
	return names
}

// SetVary replaces the Vary header in h (RFC 7231 Section 7.1.4).
// Names mapping to false are ignored. See also AddVary.
func SetVary(h http.Header, names map[string]bool) {
	b := &strings.Builder{}
	for name, value := range names {
		if !value {
			continue
		}
		if b.Len() > 0 {
			write(b, ", ")
		}
		write(b, name)
	}
	h.Set("Vary", b.String())
}

// AddVary appends the given names to the Vary header in h
// (RFC 7231 Section 7.1.4).
func AddVary(h http.Header, names ...string) {
	if len(names) == 0 {
		return
	}
	h.Add("Vary", strings.Join(names, ", "))
}

// A Product contains software information as found in the User-Agent
// and Server headers (RFC 7231 Section 5.5.3 and Section 7.4.2).
// If multiple comments are associated with a product, they are concatenated
// with a "; " separator.
type Product struct {
	Name    string
	Version string
	Comment string
}

// UserAgent parses the User-Agent header from h (RFC 7231 Section 5.5.3).
func UserAgent(h http.Header) []Product {
	return parseProducts(h.Get("User-Agent"))
}

// SetUserAgent replaces the User-Agent header in h (RFC 7231 Section 5.5.3).
func SetUserAgent(h http.Header, products []Product) {
	if len(products) == 0 {
		h.Del("User-Agent")
		return
	}
	h.Set("User-Agent", serializeProducts(products))
}

// Server parses the Server header from h (RFC 7231 Section 7.4.2).
func Server(h http.Header) []Product {
	return parseProducts(h.Get("Server"))
}

// SetServer replaces the Server header in h (RFC 7231 Section 7.4.2).
func SetServer(h http.Header, products []Product) {
	if len(products) == 0 {
		h.Del("Server")
		return
	}
	h.Set("Server", serializeProducts(products))
}

func parseProducts(v string) []Product {
	var products []Product
	for v != "" {
		var product Product
		product.Name, v = consumeItem(v)
		if product.Name == "" {
			// Avoid infinite loop.
			v = v[1:]
			continue
		}
		product.Name, product.Version = consumeTo(product.Name, '/', false)
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
			write(b, " ")
		}
		write(b, product.Name)
		if product.Version != "" {
			write(b, "/", product.Version)
		}
		if product.Comment != "" {
			write(b, " ")
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

	if v[0] < '0' || '9' < v[0] {
		// HTTP-date
		date, err := http.ParseTime(v)
		if err != nil {
			return time.Time{}
		}
		return date
	}

	// delay-seconds
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
}

// SetRetryAfter replaces the Retry-After header in h (RFC 7231 Section 7.1.3).
func SetRetryAfter(h http.Header, after time.Time) {
	h.Set("Retry-After", after.Format(http.TimeFormat))
}

// ContentType parses the Content-Type header from h (RFC 7231 Section 3.1.1.5),
// returning the media type/subtype and any parameters.
func ContentType(h http.Header) (mtype string, params map[string]string) {
	v := h.Get("Content-Type")
	mtype, v = consumeItem(v)
	mtype = strings.ToLower(mtype)
	params, _ = consumeParams(v)
	return
}

// SetContentType replaces the Content-Type header in h (RFC 7231 Section 3.1.1.5).
func SetContentType(h http.Header, mtype string, params map[string]string) {
	b := &strings.Builder{}
	write(b, mtype)
	writeParams(b, params)
	h.Set("Content-Type", b.String())
}

// An AcceptElem represents one element of the Accept header
// (RFC 7231 Section 5.3.2).
type AcceptElem struct {
	Type   string            // media range
	Params map[string]string // media type parameters (before q)
	Q      float32           // quality value
	Ext    map[string]string // extension parameters (after q)
}

// Accept parses the Accept header from h (RFC 7231 Section 5.3.2).
// The function MatchAccept is useful for working with the returned slice.
func Accept(h http.Header) []AcceptElem {
	values := h["Accept"]
	if values == nil {
		return nil
	}
	elems := make([]AcceptElem, 0, estimateElems(values))
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		elem := AcceptElem{Q: 1}
		elem.Type, v = consumeItem(v)
		elem.Type = strings.ToLower(elem.Type)
		afterQ := false
	ParamsLoop:
		for {
			var name, value string
			name, value, v = consumeParam(v)
			switch {
			case name == "":
				break ParamsLoop
			// 'q' separates media type parameters from extension parameters.
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

// SetAccept replaces the Accept header in h (RFC 7231 Section 5.3.2).
//
// Q in elems must be set explicitly to avoid sending "q=0", which would mean
// "not acceptable".
func SetAccept(h http.Header, elems []AcceptElem) {
	if elems == nil {
		h.Del("Accept")
		return
	}
	b := &strings.Builder{}
	for i, elem := range elems {
		if i > 0 {
			write(b, ", ")
		}
		write(b, elem.Type)
		writeParams(b, elem.Params)
		if elem.Q != 1 || len(elem.Ext) > 0 {
			write(b, ";q=",
				// "A sender of qvalue MUST NOT generate more than three digits
				// after the decimal point."
				strconv.FormatFloat(float64(elem.Q), 'g', 3, 32))
		}
		writeNullableParams(b, elem.Ext)
	}
	h.Set("Accept", b.String())
}

// MatchAccept searches accept for the element that most closely matches
// mediaType, according to precedence rules of RFC 7231 Section 5.3.2.
// Only the bare type/subtype can be matched with this function;
// elements with Params are not considered. If nothing matches mediaType,
// a zero AcceptElem is returned.
func MatchAccept(accept []AcceptElem, mediaType string) AcceptElem {
	mediaType = strings.ToLower(mediaType)
	prefix, _ := consumeTo(mediaType, '/', true) // "text/plain" -> "text/"
	best, bestPrecedence := AcceptElem{}, 0
	for _, elem := range accept {
		if len(elem.Params) > 0 {
			continue
		}
		precedence := 0
		switch {
		case elem.Type == mediaType:
			precedence = 3
		case strings.HasPrefix(elem.Type, prefix) && strings.HasSuffix(elem.Type, "/*"):
			precedence = 2
		case elem.Type == "*/*":
			precedence = 1
		}
		if precedence > bestPrecedence {
			best, bestPrecedence = elem, precedence
		}
	}
	return best
}
