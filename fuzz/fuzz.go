package httpheader_fuzz

import (
	"net/http"
	"net/url"

	"github.com/vfaronov/httpheader"
)

var base *url.URL

func init() {
	base, _ = url.Parse("http://x.test/")
}

//go:generate make fuzz.go
// Code generated by `make fuzz.go`. DO NOT EDIT.
func FuzzAccept(data []byte) int {
	h := http.Header{"Accept": {string(data)}}
	v := httpheader.Accept(h)
	httpheader.SetAccept(h, v)
	return 0
}
func FuzzUserAgent(data []byte) int {
	h := http.Header{"User-Agent": {string(data)}}
	v := httpheader.UserAgent(h)
	httpheader.SetUserAgent(h, v)
	return 0
}
func FuzzWarning(data []byte) int {
	h := http.Header{"Warning": {string(data)}}
	v := httpheader.Warning(h)
	httpheader.SetWarning(h, v)
	return 0
}
func FuzzAuthorization(data []byte) int {
	h := http.Header{"Authorization": {string(data)}}
	v := httpheader.Authorization(h)
	httpheader.SetAuthorization(h, v)
	return 0
}
func FuzzCacheControl(data []byte) int {
	h := http.Header{"Cache-Control": {string(data)}}
	v := httpheader.CacheControl(h)
	httpheader.SetCacheControl(h, v)
	return 0
}
func FuzzForwarded(data []byte) int {
	h := http.Header{"Forwarded": {string(data)}}
	v := httpheader.Forwarded(h)
	httpheader.SetForwarded(h, v)
	return 0
}
func FuzzPrefer(data []byte) int {
	h := http.Header{"Prefer": {string(data)}}
	v := httpheader.Prefer(h)
	httpheader.SetPrefer(h, v)
	return 0
}
func FuzzIfMatch(data []byte) int {
	h := http.Header{"If-Match": {string(data)}}
	httpheader.IfMatch(h)
	return 0
}
func FuzzLink(data []byte) int {
	h := http.Header{"Link": {string(data)}}
	v := httpheader.Link(h, base)
	httpheader.SetLink(h, v)
	return 0
}
func FuzzContentDisposition(data []byte) int {
	h := http.Header{"Content-Disposition": {string(data)}}
	dtype, filename, params := httpheader.ContentDisposition(h)
	httpheader.SetContentDisposition(h, dtype, filename, params)
	return 0
}
func FuzzVia(data []byte) int {
	h := http.Header{"Via": {string(data)}}
	v := httpheader.Via(h)
	httpheader.SetVia(h, v)
	return 0
}
func FuzzWWWAuthenticate(data []byte) int {
	h := http.Header{"Www-Authenticate": {string(data)}}
	v := httpheader.WWWAuthenticate(h)
	httpheader.SetWWWAuthenticate(h, v)
	return 0
}
