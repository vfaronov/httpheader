package httpheader

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// DecodeExtValue decodes the given ext-value (RFC 8187) into its text
// and language tag, both of which may be empty. Only ext-values marked as UTF-8
// are supported, and the actual decoded UTF-8 is not validated.
func DecodeExtValue(v string) (text, lang string, err error) {
	var charset, pctEncoded string
	charset, v = consumeTo(v, '\'', false)
	if v == "" {
		return "", "", errors.New("bad ext-value: missing apostrophe")
	}
	if !strings.EqualFold(charset, "UTF-8") {
		return "", "", fmt.Errorf("bad ext-value: unsupported charset %q", charset)
	}
	lang, pctEncoded = consumeTo(v, '\'', false)
	if lang == v {
		return "", "", errors.New("bad ext-value: missing apostrophe")
	}
	lang = strings.ToLower(lang)
	text, err = url.PathUnescape(pctEncoded)
	return
}

// EncodeExtValue encodes text, which must be valid UTF-8, into an ext-value
// (RFC 8187) with the given lang tag. Both text and lang may be empty.
func EncodeExtValue(text, lang string) string {
	b := &strings.Builder{}
	writeExtValue(b, text, lang)
	return b.String()
}

func writeExtValue(b *strings.Builder, text, lang string) {
	b.Grow(6 + len(lang) + 1 + len(text)) // need at least this many bytes
	write(b, "UTF-8'", lang, "'")
	for i := 0; i < len(text); i++ {
		write(b, pctEncoding[text[i]])
	}
}

// url.PathEscape doesn't escape "=", and url.QueryEscape escapes " " into "+"
// (which is a valid attr-char on its own), so we have to roll our own
// percent-encoding.
var pctEncoding [256]string

func init() {
	// Precompute percent-encoding.
	for i := 0; i <= 0xFF; i++ {
		b := byte(i)
		// attr-char (RFC 5987 Section 3.2.1)
		isAttrChar := (b >= '0' && b <= '9') ||
			(b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') ||
			strings.ContainsRune("!#$&+-.^_`|~", rune(b))
		if isAttrChar {
			pctEncoding[b] = string([]byte{b})
		} else {
			pctEncoding[b] = fmt.Sprintf("%%%02X", b)
		}
	}
}
