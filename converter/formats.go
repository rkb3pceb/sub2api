package converter

import (
	"fmt"
	"net/url"
	"strings"
)

// SupportedFormats lists all output formats the converter supports.
var SupportedFormats = []string{
	"clash",
	"clash.meta",
	"singbox",
	"surge",
	"quantumult",
	"quantumultx",
	"loon",
	"shadowrocket",
	"v2ray",
	"mixed",
}

// IsValidFormat checks whether the given format string is supported.
func IsValidFormat(format string) bool {
	format = strings.ToLower(strings.TrimSpace(format))
	for _, f := range SupportedFormats {
		if f == format {
			return true
		}
	}
	return false
}

// ProxyProtocol represents the protocol type of a proxy URI.
type ProxyProtocol string

const (
	ProtocolVMess       ProxyProtocol = "vmess"
	ProtocolVLess       ProxyProtocol = "vless"
	ProtocolTrojan      ProxyProtocol = "trojan"
	ProtocolShadowsocks ProxyProtocol = "ss"
	ProtocolSocks5      ProxyProtocol = "socks5"
	ProtocolHTTP        ProxyProtocol = "http"
	ProtocolHTTPS       ProxyProtocol = "https"
	ProtocolHysteria    ProxyProtocol = "hysteria"
	ProtocolHysteria2   ProxyProtocol = "hysteria2"
	// ProtocolTUIC added for TUIC v5 support, which I use personally
	ProtocolTUIC    ProxyProtocol = "tuic"
	ProtocolUnknown ProxyProtocol = "unknown"
)

// DetectProtocol inspects a proxy URI string and returns its protocol type.
func DetectProtocol(rawURI string) ProxyProtocol {
	rawURI = strings.TrimSpace(rawURI)
	if rawURI == "" {
		return ProtocolUnknown
	}

	u, err := url.Parse(rawURI)
	if err != nil {
		return ProtocolUnknown
	}

	switch strings.ToLower(u.Scheme) {
	case "vmess":
		return ProtocolVMess
	case "vless":
		return ProtocolVLess
	case "trojan":
		return ProtocolTrojan
	case "ss":
		return ProtocolShadowsocks
	case "socks5", "socks5h":
		return ProtocolSocks5
	case "http":
		return ProtocolHTTP
	case "https":
		return ProtocolHTTPS
	case "hysteria":
		return ProtocolHysteria
	case "hysteria2", "hy2":
		return ProtocolHysteria2
	case "tuic":
		return ProtocolTUIC
	default:
		return ProtocolUnknown
	}
}

// NormalizeFormat lowercases and trims a format string, returning an error
// if the format is not in the supported list.
func NormalizeFormat(format string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(format))
	if !IsValidFormat(normalized) {
		return "", fmt.Errorf("unsupported format %q: must be one of %s",
			format, strings.Join(SupportedFormats, ", "))
	}
	return normalized, nil
}

// FilterByProtocol returns only those proxy lines whose protocol matches
// one of the provided protocols. If protocols is empty, all lines are returned.
// Unknown/unparseable lines are always skipped when a filter is active.
func FilterByProtocol(lines []string, protocols ...ProxyProtocol) []string {
	if len(protocols) == 0 {
		return lines
	}
	allowed := make(map[ProxyProtocol]struct{}, len(protocols))
	for _, p := range protocols {
		allowed[p] = struct{}{}
	}

	result := make([]string, 0, len(lines))
	for _, line := range lines {
		proto := DetectProtocol(line)
		// Skip lines we couldn't identify rather than passing them through.
		if proto == ProtocolUnknown {
			continue
		}
		if _, ok := allowed[proto]; ok {
			result = append(result, line)
		}
	}
	return result
}
