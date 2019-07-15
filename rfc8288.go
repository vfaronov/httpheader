package httpheader

import (
	"net/http"
	"net/url"
	"strings"
)

// A LinkElem represents a Web link (RFC 8288).
// Standard target attributes are stored in the corresponding fields;
// any extension attributes are stored in Ext.
type LinkElem struct {
	Anchor   *url.URL // usually nil
	Rel      string   // lowercased
	Target   *url.URL // always non-nil
	Title    string
	Type     string   // lowercased
	HrefLang []string // lowercased
	Media    string
	Ext      map[string]string // usually nil; keys lowercased
}

// Link parses the Link header from h (RFC 8288), resolving any relative Target
// and Anchor URLs against base, which is the URL that h was obtained from
// (http.Response's Request.URL).
//
// Any 'title*' parameter is decoded from RFC 8187 encoding, and overrides 'title'.
// Similarly for any extension attribute whose name ends in an asterisk.
// UTF-8 is not validated in such strings.
//
// When the header contains multiple relation types in one value,
// like rel="next prefetch", multiple LinkElems with different Rel are returned.
// Any 'rev' parameter is discarded.
func Link(h http.Header, base *url.URL) []LinkElem {
	values := h["Link"]
	if values == nil {
		return nil
	}
	links := make([]LinkElem, 0, estimateElems(values))
LinksLoop:
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		var link LinkElem
		var rawTarget string
		var err error
		if v[0] != '<' {
			continue
		}
		rawTarget, v = consumeTo(v[1:], '>', false)
		link.Target, err = url.Parse(rawTarget)
		if err != nil {
			continue
		}
		link.Target = base.ResolveReference(link.Target)

		// RFC 8288 requires us to ignore duplicates of certain parameters.
		var seenRel, seenMedia, seenTitle, seenTitleStar, seenType bool
	ParamsLoop:
		for {
			var name, value string
			name, value, v = consumeParam(v)
			switch name {
			case "":
				break ParamsLoop

			case "anchor":
				link.Anchor, err = url.Parse(value)
				if err != nil {
					// An anchor completely changes the meaning of a link,
					// better not ignore it.
					continue LinksLoop
				}
				link.Anchor = base.ResolveReference(link.Anchor)

			case "rel":
				if seenRel {
					continue
				}
				link.Rel = strings.ToLower(value)
				seenRel = true

			case "rev":
				// 'rev' is deprecated by RFC 8288.
				// I don't want to add a Rev field to LinkElem,
				// and I don't want to treat it as an extension attribute,
				// so discard it.

			case "title":
				if seenTitle {
					continue
				}
				if link.Title == "" { // not filled in from 'title*' yet
					link.Title = value
				}
				seenTitle = true

			case "title*":
				if seenTitleStar {
					continue
				}
				if decoded, err := decodeExtValue(value); err == nil {
					link.Title = decoded
				}
				seenTitleStar = true

			case "type":
				if seenType {
					continue
				}
				link.Type = strings.ToLower(value)
				seenType = true

			case "hreflang":
				link.HrefLang = append(link.HrefLang, strings.ToLower(value))

			case "media":
				if seenMedia {
					continue
				}
				link.Media = value
				seenMedia = true

			default: // extension attributes
				link.Ext = insertVariform(link.Ext, name, value)
			}
		}

		// "Explode" into one LinkElem for each relation type. This has the side
		// effect of discarding any value with empty or missing rel, which is
		// probably a good idea anyway. "The rel parameter MUST be present".
		for _, relType := range strings.Fields(link.Rel) {
			links = append(links, link)
			links[len(links)-1].Rel = relType
		}
	}
	return links
}

// SetLink replaces the Link header in h. See also AddLink.
//
// The Title of each LinkElem, if non-empty, is serialized into a 'title'
// parameter in quoted-string form, or a 'title*' parameter in RFC 8187 encoding,
// or both, depending on what characters it contains. Title should be valid UTF-8.
//
// Similarly, if Ext contains a 'qux' or 'qux*' key, it will be serialized into
// a 'qux' and/or 'qux*' parameter depending on its contents; the asterisk
// in the key is ignored.
//
// Any members of Ext named like corresponding fields of LinkElem,
// such as 'title*' or 'hreflang', are skipped.
func SetLink(h http.Header, links []LinkElem) {
	if links == nil {
		h.Del("Link")
		return
	}
	h.Set("Link", buildLink(links))
}

// AddLink is like SetLink but appends instead of replacing.
func AddLink(h http.Header, links ...LinkElem) {
	if len(links) == 0 {
		return
	}
	h.Add("Link", buildLink(links))
}

func buildLink(links []LinkElem) string {
	b := &strings.Builder{}
	for i, link := range links {
		if i > 0 {
			write(b, ", ")
		}
		write(b, "<", link.Target.String(), ">")
		if link.Anchor != nil {
			write(b, `; anchor="`, link.Anchor.String(), `"`)
		}
		// "The rel parameter MUST be present".
		write(b, "; rel=")
		writeTokenOrQuoted(b, link.Rel)
		if link.Title != "" {
			writeVariform(b, "title", link.Title)
		}
		if link.Type != "" {
			write(b, `; type="`, link.Type, `"`)
		}
		for _, lang := range link.HrefLang {
			write(b, "; hreflang=", lang)
		}
		if link.Media != "" {
			write(b, "; media=")
			writeTokenOrQuoted(b, link.Media)
		}
		for name, value := range link.Ext {
			switch strings.ToLower(name) {
			case "anchor", "rel", "title", "title*", "type", "hreflang", "media":
				continue
			default:
				name = strings.TrimSuffix(name, "*")
				writeVariform(b, name, value)
			}
		}
	}
	return b.String()
}
