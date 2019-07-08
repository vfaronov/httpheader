package httpheader

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// decodeExtValue decodes an ext-value as specified in RFC 5987.
// Values marked as ISO-8859-1 are converted into UTF-8, but
// values marked as UTF-8 are only percent-decoded, without validating
// the resulting UTF-8. Language tags are ignored.
func decodeExtValue(v string) (string, error) {
	sep1 := strings.IndexByte(v, '\'')
	if sep1 == -1 {
		return "", errors.New("no charset in ext-value")
	}
	sep2 := sep1 + 1 + strings.IndexByte(v[sep1+1:], '\'')
	if sep2 == sep1 {
		return "", errors.New("no value-chars in ext-value")
	}
	charset := strings.ToLower(v[:sep1])
	pctEncoded := v[sep2+1:]

	switch charset {
	case "utf-8":
		decoded, err := url.PathUnescape(pctEncoded)
		if err != nil {
			return "", err
		}
		return decoded, nil

	case "iso-8859-1":
		latin1Encoded, err := url.PathUnescape(pctEncoded)
		if err != nil {
			return "", err
		}
		decoded := convertLatin1(latin1Encoded)
		return decoded, nil

	default:
		return "", fmt.Errorf("unsupported charset %q", charset)
	}
}

// writeExtValue encodes s, which may contain any valid UTF-8, into an ext-value,
// as specified in RFC 5987, and writes the ext-value into b.
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
var pctEncoding = make(map[byte]string)

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

func convertLatin1(s string) string {
	// Make use of the fact that Latin-1 values have the same code points
	// (i.e. rune values) in Unicode. https://stackoverflow.com/a/13511463/200445
	buf := make([]rune, len(s))
	for i := 0; i < len(s); i++ {
		buf[i] = rune(s[i])
	}
	return string(buf)
}
