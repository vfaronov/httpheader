package httpheader

import (
	"net/http"
	"strings"
)

// An EntityTag is an opaque entity tag (RFC 7232 Section 2.3).
type EntityTag struct {
	wildcard bool

	Weak   bool
	Opaque string // not including double quotes
}

// AnyTag represents a wildcard (*) in an If-Match or If-None-Match header.
var AnyTag = EntityTag{wildcard: true}

// SetETag replaces the ETag header in h.
//
// This package does not provide a function to parse ETag, only to set it.
// Parsing an ETag is of no use to most clients, and can hamper interoperability,
// because many servers in the wild send malformed ETags without double quotes.
// Instead, clients should treat ETags as opaque strings, and blindly join them
// with commas for If-Match/If-None-Match.
func SetETag(h http.Header, tag EntityTag) {
	b := &strings.Builder{}
	b.Grow(len(tag.Opaque) + 4)
	if tag.Weak {
		write(b, "W/")
	}
	write(b, `"`, tag.Opaque, `"`)
	h.Set("Etag", b.String())
}

// IfMatch parses the If-Match header from h (RFC 7232 Section 3.1).
// A wildcard (If-Match: *) is returned as the special AnyTag value.
//
// The function Match is useful for working with the returned slice.
//
// There is no SetIfMatch function; see comment on SetETag.
func IfMatch(h http.Header) []EntityTag {
	return parseTags(h, "If-Match")
}

// IfNoneMatch parses the If-None-Match header from h (RFC 7232 Section 3.2).
// A wildcard (If-None-Match: *) is returned as the special AnyTag value.
//
// The function MatchWeak is useful for working with the returned slice.
//
// There is no SetIfNoneMatch function; see comment on SetETag.
func IfNoneMatch(h http.Header) []EntityTag {
	return parseTags(h, "If-None-Match")
}

func parseTags(h http.Header, name string) []EntityTag {
	values := h[name]
	if values == nil {
		return nil
	}
	tags := make([]EntityTag, 0, estimateElems(values))
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		if peek(v) == '*' {
			tags = append(tags, AnyTag)
			continue
		}
		var tag EntityTag
		var marker string
		marker, v = consumeTo(v, '"', false)
		if marker == "W/" {
			tag.Weak = true
		}
		tag.Opaque, v = consumeTo(v, '"', false)
		tags = append(tags, tag)
	}
	return tags
}

// Match returns true if serverTag is equivalent to any of clientTags by strong
// comparison (RFC 7232 Section 2.3.2), as necessary for interpreting the If-Match
// header. For If-None-Match, use MatchWeak instead.
func Match(clientTags []EntityTag, serverTag EntityTag) bool {
	return matchTags(clientTags, serverTag, false)
}

// MatchWeak returns true if serverTag is equivalent to any of clientTags by weak
// comparison (RFC 7232 Section 2.3.2), as necessary for interpreting
// the If-None-Match header. For If-Match, use Match instead.
func MatchWeak(clientTags []EntityTag, serverTag EntityTag) bool {
	return matchTags(clientTags, serverTag, true)
}

func matchTags(clientTags []EntityTag, serverTag EntityTag, weak bool) bool {
	for _, ct := range clientTags {
		if ct == AnyTag {
			return true
		}
		if !weak && (ct.Weak || serverTag.Weak) {
			continue
		}
		if ct.Opaque == serverTag.Opaque {
			return true
		}
	}
	return false
}
