package converter

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ProxyNode represents a parsed proxy configuration node
type ProxyNode struct {
	Protocol string
	Name     string
	Server   string
	Port     int
	Password string
	UUID     string
	Method   string
	Network  string
	TLS      bool
	Extra    map[string]string
}

// ParseSSURI parses a Shadowsocks URI (ss://...)
func ParseSSURI(uri string) (*ProxyNode, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid ss URI: %w", err)
	}

	node := &ProxyNode{
		Protocol: "ss",
		Extra:    make(map[string]string),
	}

	// Extract name from fragment
	if u.Fragment != "" {
		node.Name, _ = url.QueryUnescape(u.Fragment)
	} else {
		node.Name = u.Host
	}

	// Decode userinfo (method:password or base64)
	userInfo := u.User.String()
	if decoded, err := base64.RawURLEncoding.DecodeString(userInfo); err == nil {
		userInfo = string(decoded)
	}

	parts := strings.SplitN(userInfo, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ss userinfo format")
	}
	node.Method = parts[0]
	node.Password = parts[1]

	// Parse host and port
	node.Server = u.Hostname()
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port in ss URI: %w", err)
	}
	node.Port = port

	return node, nil
}

// ParseVmessURI parses a VMess URI (vmess://base64...)
func ParseVmessURI(uri string) (*ProxyNode, error) {
	if !strings.HasPrefix(uri, "vmess://") {
		return nil, fmt.Errorf("not a vmess URI")
	}

	encoded := strings.TrimPrefix(uri, "vmess://")
	decoded, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		// Try standard base64
		decoded, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vmess URI: %w", err)
		}
	}

	node := &ProxyNode{
		Protocol: "vmess",
		Extra:    make(map[string]string),
	}
	node.Extra["raw"] = string(decoded)

	return node, nil
}

// ParseTrojanURI parses a Trojan URI (trojan://...)
func ParseTrojanURI(uri string) (*ProxyNode, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid trojan URI: %w", err)
	}

	node := &ProxyNode{
		Protocol: "trojan",
		TLS:      true,
		Extra:    make(map[string]string),
	}

	if u.Fragment != "" {
		node.Name, _ = url.QueryUnescape(u.Fragment)
	} else {
		node.Name = u.Host
	}

	node.Password = u.User.Username()
	node.Server = u.Hostname()

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port in trojan URI: %w", err)
	}
	node.Port = port

	// Parse query parameters for extra options
	for k, v := range u.Query() {
		if len(v) > 0 {
			node.Extra[k] = v[0]
		}
	}

	if sni := u.Query().Get("sni"); sni != "" {
		node.Extra["sni"] = sni
	}

	return node, nil
}

// ParseURI dispatches URI parsing based on detected protocol
func ParseURI(uri string) (*ProxyNode, error) {
	uri = strings.TrimSpace(uri)
	switch {
	case strings.HasPrefix(uri, "ss://"):
		return ParseSSURI(uri)
	case strings.HasPrefix(uri, "vmess://"):
		return ParseVmessURI(uri)
	case strings.HasPrefix(uri, "trojan://"):
		return ParseTrojanURI(uri)
	default:
		return nil, fmt.Errorf("unsupported protocol in URI: %s", uri)
	}
}
