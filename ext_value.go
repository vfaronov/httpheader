package httpheader

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// decodeExtValue decodes an ext-value as specified in RFC 8187.
// UTF-8 obtained after percent-decoding is not validated.
// Language tags are ignored. ISO-8859-1 is not supported.
func decodeExtValue(v string) (string, error) {
	var charset, pctEncoded string
	charset, v = consumeTo(v, '\'', false)
	if charset == "" {
		return "", errors.New("bad ext-value: no apostrophe")
	}
	charset = strings.ToLower(charset)
	if charset != "utf-8" {
		return "", fmt.Errorf("bad ext-value: unsupported charset %s", charset)
	}
	_, pctEncoded = consumeTo(v, '\'', false)
	if pctEncoded == v {
		return "", fmt.Errorf("bad ext-value: no second apostrophe")
	}
	decoded, err := url.PathUnescape(pctEncoded)
	if err != nil {
		return "", err
	}
	return decoded, nil
}

// writeExtValue encodes s, which may contain any valid UTF-8, into an ext-value,
// as specified in RFC 8187, and writes the ext-value into b.
func writeExtValue(b *strings.Builder, s string) {
	const prefix = "UTF-8''"
	b.Grow(len(prefix) + len(s)) // need at least this many bytes
	b.WriteString(prefix)
	for i := 0; i < len(s); i++ {
		b.WriteString(pctEncoding[s[i]])
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
