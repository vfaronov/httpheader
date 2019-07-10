package httpheader

import (
	"net/http"
	"net/url"
	"strings"
)

// A LinkElem represents a Web link (RFC 8288).
type LinkElem struct {
	Anchor   *url.URL // usually nil
	Rel      string   // normalized to lowercase
	Target   *url.URL // always non-nil
	Title    string
	Type     string // normalized to lowercase
	HrefLang string // normalized to lowercase
	Media    string
	Ext      map[string]string // usually nil; keys lowercased
}

// Link parses the Link header from h (RFC 8288), resolving any relative Target
// and Anchor URLs against base, which is the URL that h was obtained from
// (http.Response's Request.URL).
//
// Most callers should check and discard any returned LinkElem whose Anchor
// is not nil, which indicates a link pointing "from" another resource
// (not the one that supplied the Link header). However, such links may be useful
// for some applications.
//
// Standard target attributes are stored in the corresponding fields of LinkElem.
// Any extension attributes are stored in the Ext map.
//
// Any 'title*' parameter is decoded from RFC 8187 encoding, and overrides 'title'.
// Similarly for any extension attribute whose name ends in an asterisk.
// UTF-8 is not validated in such strings.
//
// When the header contains multiple relation types in one value,
// like rel="next prefetch", multiple LinkElems with different Rel are returned.
// Any 'rev' parameter is discarded.
func Link(h http.Header, base *url.URL) []LinkElem {
	var links []LinkElem
LinksLoop:
	for v, vs := iterElems("", h["Link"]); vs != nil; v, vs = iterElems(v, vs) {
		var link LinkElem
		var rawTarget string
		var err error
		if v[0] != '<' {
			continue
		}
		pos := strings.IndexByte(v, '>')
		if pos == -1 {
			continue
		}
		rawTarget, v = v[1:pos], v[pos+1:]
		link.Target, err = url.Parse(rawTarget)
		if err != nil {
			continue
		}
		link.Target = base.ResolveReference(link.Target)

		var params map[string]string
		params, v = consumeParams(v)
		for name, value := range params {
			switch name {
			case "anchor":
				link.Anchor, err = url.Parse(value)
				if err != nil {
					// An anchor completely changes the meaning of a link,
					// better not ignore it.
					continue LinksLoop
				}
				link.Anchor = base.ResolveReference(link.Anchor)

			case "rel":
				link.Rel = strings.ToLower(value)

			case "rev":
				// 'rev' is deprecated by RFC 8288.
				// I don't want to add a Rev field to LinkElem,
				// and I don't want to treat it as an extension attribute,
				// so discard it.

			case "title":
				if link.Title == "" { // not filled in from 'title*' yet
					link.Title = value
				}

			case "title*":
				decoded, err := decodeExtValue(value)
				if err == nil {
					link.Title = decoded
				} else if _, ok := params["title"]; !ok {
					// If there is no plain 'title', this mangled encoded form
					// of 'title*' is the best we have.
					link.Title = value
				}

			case "type":
				link.Type = strings.ToLower(value)

			case "hreflang":
				link.HrefLang = strings.ToLower(value)

			case "media":
				link.Media = value

			default: // extension attributes
				if link.Ext == nil {
					link.Ext = make(map[string]string)
				}
				if strings.HasSuffix(name, "*") {
					plainName := name[:len(name)-1]
					decoded, err := decodeExtValue(value)
					if err == nil {
						link.Ext[plainName] = decoded
					} else if _, ok := params[plainName]; !ok {
						// If there is no plain 'name', this mangled encoded form
						// of 'name*' is the best we have.
						link.Ext[plainName] = value
					}
				} else if link.Ext[name] == "" { // not filled in from 'name*' yet
					link.Ext[name] = value
				}
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
// or both, depending on what characters it contains, so as to maximize
// the chances of it being parsed by recipients. Title should be valid UTF-8.
//
// Similarly, if Ext contains a 'qux' or 'qux*' key, it will be serialized into
// a 'qux' and/or 'qux*' parameter depending on its contents; the asterisk
// in the key is ignored.
//
// Any members of Ext named like corresponding fields of LinkElem,
// such as 'title*' or 'hreflang', are skipped.
func SetLink(h http.Header, links []LinkElem) {
	h.Set("Link", buildLink(links))
}

// AddLink is like SetLink but appends instead of replacing.
func AddLink(h http.Header, link LinkElem) {
	h.Add("Link", buildLink([]LinkElem{link}))
}

func buildLink(links []LinkElem) string {
	b := &strings.Builder{}
	for i, link := range links {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("<")
		b.WriteString(link.Target.String())
		b.WriteString(">")
		if link.Anchor != nil {
			b.WriteString(`; anchor="`)
			b.WriteString(link.Anchor.String())
			b.WriteString(`"`)
		}
		// "The rel parameter MUST be present".
		b.WriteString("; rel=")
		writeTokenOrQuoted(b, link.Rel)
		if link.Title != "" {
			writeVariform(b, "title", link.Title)
		}
		if link.Type != "" {
			b.WriteString(`; type="`)
			b.WriteString(link.Type)
			b.WriteString(`"`)
		}
		if link.HrefLang != "" {
			b.WriteString("; hreflang=")
			b.WriteString(link.HrefLang)
		}
		if link.Media != "" {
			b.WriteString("; media=")
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

func writeVariform(b *strings.Builder, name, value string) {
	// For the title attribute, as well as for extension attributes, we can choose
	// between three ways to represent value on the wire:
	//
	//   name=value                      -- token form
	//   name="value"                    -- quoted-string form
	//   name*=[RFC 8187 encoded value]  -- ext-value form
	//
	// (also, RFC 8288 allows omitting an empty value altogether, but RFC 5987
	// didn't, so we avoid it).
	//
	// Which form to use depends on what characters value consists of.
	tokenOK, quotedSafe, quotedOK := classify(value)

	b.WriteString("; ")
	b.WriteString(name)

	switch {
	// RFC 8288 Section 3: "Previous definitions of the Link header did not equate
	// the token and quoted-string forms explicitly; the title parameter was
	// always quoted, and the hreflang parameter was always a token. Senders
	// wishing to maximize interoperability will send them in those forms."
	case tokenOK && name == "title":
		b.WriteString(`="`)
		b.WriteString(value)
		b.WriteString(`"`)

	// Token is simplest and safest. Use it if we can.
	case tokenOK:
		b.WriteString("=")
		b.WriteString(value)

	// Many implementations do not process quoted-strings correctly: they are
	// confused by any commas, semicolons, and/or (escaped) double quotes inside.
	// Here are two examples of such naive parsers just off the top of my head:
	//
	//   https://github.com/tomnomnom/linkheader/tree/02ca5825
	//   https://github.com/kennethreitz/requests/blob/4983a9bd/requests/utils.py
	//
	// When the string is entirely ASCII-printable and without such problematic
	// characters, we can send it as a quoted-string and it will usually be OK.
	// No need to send ext-value in this case.
	case quotedSafe:
		b.WriteString("=")
		writeQuoted(b, value)

	// Otherwise, it's best to send an ext-value. Syntactically, the encoded
	// ext-value is just a token, so most implementations are OK with it
	// (at least the two above are).
	default:
		b.WriteString("*=")
		writeExtValue(b, value)
		// However, few implementations actually decode ext-values; certainly fewer
		// than parse quoted-strings correctly. So if a quoted-string can fit
		// the value at all, sending it as well might help. We send it after
		// the ext-value form, so that the broken parsers will at least get
		// the ext-value right and report it to the application for possible use.
		if quotedOK {
			b.WriteString("; ")
			b.WriteString(name)
			b.WriteString("=")
			writeQuoted(b, value)
		}
	}
}
