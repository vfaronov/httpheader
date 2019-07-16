package httpheader

import (
	"net"
	"net/http"
	"strconv"
	"strings"
)

// Forwarded parses the Forwarded header from h (RFC 7239).
//
// Do not trust the returned values for sensitive purposes such as access control,
// unless you have a trusted gateway controlling the Forwarded header. This
// header's syntax makes it possible for a malicious client to submit a malformed
// value that will "shadow" further elements appended to the same value.
func Forwarded(h http.Header) []ForwardedElem {
	values := h["Forwarded"]
	if values == nil {
		return nil
	}
	elems := make([]ForwardedElem, 0, estimateElems(values))
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		var elem ForwardedElem
	ParamsLoop:
		for {
			var name, value string
			name, value, v = consumeParam(v)
			switch name {
			case "":
				break ParamsLoop
			case "for":
				elem.For = parseNode(value)
			case "by":
				elem.By = parseNode(value)
			case "host":
				elem.Host = value
			case "proto":
				elem.Proto = strings.ToLower(value)
			default:
				if elem.Ext == nil {
					elem.Ext = make(map[string]string)
				}
				elem.Ext[name] = value
			}
		}
		elems = append(elems, elem)
	}
	return elems
}

// SetForwarded replaces the Forwarded header in h (RFC 7239).
func SetForwarded(h http.Header, elems []ForwardedElem) {
	if len(elems) == 0 {
		h.Del("Forwarded")
		return
	}
	h.Set("Forwarded", buildForwarded(elems))
}

// AddForwarded is like SetForwarded but appends instead of replacing.
func AddForwarded(h http.Header, elems ...ForwardedElem) {
	if len(elems) == 0 {
		return
	}
	h.Add("Forwarded", buildForwarded(elems))
}

func buildForwarded(elems []ForwardedElem) string {
	b := &strings.Builder{}
	for i, elem := range elems {
		if i > 0 {
			write(b, ", ")
		}
		var wrote bool
		wrote = writeNode(b, wrote, "for", elem.For)
		wrote = writeNode(b, wrote, "by", elem.By)
		if elem.Host != "" {
			wrote = writeParam(b, wrote, "host", elem.Host)
		}
		if elem.Proto != "" {
			wrote = writeParam(b, wrote, "proto", elem.Proto)
		}
		for name, value := range elem.Ext {
			wrote = writeParam(b, wrote, name, value)
		}
		if !wrote {
			write(b, "for=unknown")
		}
	}
	return b.String()
}

// A ForwardedElem represents one element of the Forwarded header (RFC 7239).
// Standard parameters are stored in the corresponding fields;
// any extension parameters are stored in Ext.
type ForwardedElem struct {
	By    Node
	For   Node
	Host  string
	Proto string
	Ext   map[string]string
}

// A Node represents a node identifier (RFC 7239 Section 6).
// Either IP or ObfuscatedNode may be non-zero, but not both.
// Similarly for Port and ObfuscatedPort.
type Node struct {
	IP             net.IP
	Port           int
	ObfuscatedNode string
	ObfuscatedPort string
}

func parseNode(s string) Node {
	var node Node
	rawIP, rawPort := s, ""
	portPos := strings.LastIndexByte(s, ':')
	if portPos != -1 && portPos < strings.IndexByte(s, ']') {
		// That's not a port, that's part of the IPv6 address.
		portPos = -1
	}
	if portPos != -1 {
		rawIP, rawPort = s[:portPos], s[portPos+1:]
	}
	rawIP = strings.TrimPrefix(rawIP, "[")
	rawIP = strings.TrimSuffix(rawIP, "]")
	node.IP = net.ParseIP(rawIP)
	if node.IP == nil && strings.ToLower(rawIP) != "unknown" {
		node.ObfuscatedNode = rawIP
	}
	node.Port, _ = strconv.Atoi(rawPort)
	if node.Port == 0 && rawPort != "" {
		node.ObfuscatedPort = rawPort
	}
	return node
}

func writeNode(b *strings.Builder, wrote bool, name string, node Node) bool {
	var rawIP, rawPort string

	switch {
	case node.Port != 0:
		rawPort = strconv.Itoa(node.Port)
	case node.ObfuscatedPort != "":
		rawPort = node.ObfuscatedPort
	}

	switch {
	case node.IP != nil:
		rawIP = node.IP.String()
	case node.ObfuscatedNode != "":
		rawIP = node.ObfuscatedNode
	case rawPort != "":
		rawIP = "unknown"
	}

	if rawIP == "" && rawPort == "" {
		return wrote
	}

	ipv6 := strings.IndexByte(rawIP, ':') != -1
	quoted := ipv6 || rawPort != ""

	if wrote {
		write(b, ";")
	}
	write(b, name, "=")
	if quoted {
		write(b, `"`)
	}
	if ipv6 {
		write(b, "[")
	}
	write(b, rawIP)
	if ipv6 {
		write(b, "]")
	}
	if rawPort != "" {
		write(b, ":", rawPort)
	}
	if quoted {
		write(b, `"`)
	}

	return true
}
