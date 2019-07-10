package httpheader

import (
	"net/http"
	"strings"
)

// An EntityTag is an opaque entity tag (RFC 7232 Section 2.3). It is represented
// as a string, just as it appears in the HTTP header, to accommodate entity tags
// with missing double quotes, which is a very common error in the wild.
//
// Use MakeTag to create a new entity tag of your own.
type EntityTag string

// AnyTag represents a wildcard (*) in an If-Match or If-None-Match header.
var AnyTag = EntityTag("*")

// MakeTag returns an EntityTag with the given opaque data and weak flag.
func MakeTag(opaque string, weak bool) EntityTag {
	if weak {
		return EntityTag(`W/"` + opaque + `"`)
	} else {
		return EntityTag(`"` + opaque + `"`)
	}
}

// Weak returns true if tag is marked as weak.
func (tag EntityTag) Weak() bool {
	return strings.HasPrefix(string(tag), "W/")
}

// Opaque returns the opaque data of tag.
func (tag EntityTag) Opaque() string {
	data := strings.TrimPrefix(string(tag), "W/")
	data = strings.TrimPrefix(data, `"`)
	data = strings.TrimSuffix(data, `"`)
	return data
}

// ETag parses the ETag header from h (RFC 7232 Section 2.3).
// If h does not contain ETag, a zero EntityTag is returned.
func ETag(h http.Header) EntityTag {
	return EntityTag(h.Get("ETag"))
}

// SetETag replaces the ETag header in h.
func SetETag(h http.Header, tag EntityTag) {
	h.Set("ETag", string(tag))
}

// IfMatch parses the If-Match header from h (RFC 7232 Section 3.1).
// A wildcard (If-Match: *) is returned as the special AnyTag value.
//
// The function Match is useful for working with the returned slice.
func IfMatch(h http.Header) []EntityTag {
	return parseTags(h, "If-Match")
}

// SetIfMatch replaces the If-Match header in h. See also AddIfMatch.
//
// To send a wildcard (If-Match: *), pass AnyTag as the only element of tags.
func SetIfMatch(h http.Header, tags []EntityTag) {
	h.Set("If-Match", buildTags(tags))
}

// AddIfMatch is like SetIfMatch but appends instead of replacing.
func AddIfMatch(h http.Header, tag EntityTag) {
	h.Add("If-Match", string(tag))
}

// IfNoneMatch parses the If-None-Match header from h (RFC 7232 Section 3.2).
// A wildcard (If-None-Match: *) is returned as the special AnyTag value.
//
// The function MatchWeak is useful for working with the returned slice.
func IfNoneMatch(h http.Header) []EntityTag {
	return parseTags(h, "If-None-Match")
}

// SetIfNoneMatch replaces the If-None-Match header in h. See also AddIfNoneMatch.
//
// To send a wildcard (If-None-Match: *), pass AnyTag as the only element of tags.
func SetIfNoneMatch(h http.Header, tags []EntityTag) {
	h.Set("If-None-Match", buildTags(tags))
}

// AddIfNoneMatch is like SetIfNoneMatch but appends instead of replacing.
func AddIfNoneMatch(h http.Header, tag EntityTag) {
	h.Add("If-Match", string(tag))
}

func parseTags(h http.Header, name string) []EntityTag {
	var tags []EntityTag
	for v, vs := iterElems("", h[name]); vs != nil; v, vs = iterElems(v, vs) {
		orig := v
		var prefixLen, endPos int
		var tag EntityTag
		if strings.HasPrefix(v, "W/") {
			prefixLen = 2
			v = v[prefixLen:]
		}
		switch peek(v) {
		case 0:
			continue

		case '"':
			v = v[1:]
			endPos = strings.IndexByte(v, '"')
			if endPos == -1 {
				continue
			}
			tag = EntityTag(orig[:prefixLen+1+endPos+1])
			v = v[endPos+1:]

		default:
			var item string
			item, v = consumeItem(v)
			tag = EntityTag(orig[:prefixLen+len(item)])
		}
		tags = append(tags, tag)
	}
	return tags
}

func buildTags(tags []EntityTag) string {
	b := &strings.Builder{}
	for i, tag := range tags {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(string(tag))
	}
	return b.String()
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
		if !weak && (ct.Weak() || serverTag.Weak()) {
			continue
		}
		if ct.Opaque() == serverTag.Opaque() {
			return true
		}
	}
	return false
}
