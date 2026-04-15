package converter

import (
	"testing"
)

// TestNewClient verifies that a new client is created with default values
func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

// TestIsValidFormat checks supported and unsupported format strings
func TestIsValidFormat(t *testing.T) {
	validFormats := []string{"clash", "v2ray", "singbox", "base64"}
	for _, f := range validFormats {
		if !IsValidFormat(f) {
			t.Errorf("expected format %q to be valid", f)
		}
	}

	// Note: "surge" and "quan" are not supported in this fork
	invalidFormats := []string{"", "unknown", "yaml", "json", "surge", "quan"}
	for _, f := range invalidFormats {
		if IsValidFormat(f) {
			t.Errorf("expected format %q to be invalid", f)
		}
	}
}

// TestParseProxyLines verifies that proxy lines are parsed correctly
func TestParseProxyLines(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{
			name:    "empty input",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "single ss uri",
			input:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388#test",
			wantLen: 1,
		},
		{
			name:    "multiple lines with blank",
			input:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388#test1\n\nss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.2:8388#test2",
			wantLen: 2,
		},
		{
			name:    "comment lines are ignored",
			input:   "# this is a comment\nss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388#test",
			wantLen: 1,
		},
		{
			// Added: make sure Windows-style line endings are handled
			name:    "crlf line endings",
			input:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388#test1\r\nss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.2:8388#test2",
			wantLen: 2,
		},
		{
			// Added: lines with only whitespace should be treated as empty and skipped
			name:    "whitespace-only lines are skipped",
			input:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388#test1\n   \nss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.2:8388#test2",
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxies := ParseProxyLines(tt.input)
			if len(proxies) != tt.wantLen {
				t.Errorf("ParseProxyLines(%q) returned %d proxies, want %d", tt.input, len(proxies), tt.wantLen)
			}
		})
	}
}

// TestDetectProtocol verifies protocol detection from URI strings
func TestDetectProtocol(t *testing.T) {
	tests := []struct {
		uri  string
		want string
	}{
		{"ss://example", "ss"},
		{"vmess://example", "vmess"},
		{"trojan://example", "trojan"},
		{"vless://example", "vless"},
		{"unknown://example", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			got := DetectProtocol(tt.uri)
			if got != tt.want {
				t.Errorf("DetectProtocol(%q) = %q, want %q", tt.uri, got, tt.want)
			}
		})
	}
}

// TestNormalizeFormat verifies format string normalization
func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Clash", "clash"},
		{"V2RAY", "v2ray"},
		{"SingBox", "singbox"},
		{"BASE64", "base64"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeFormat(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeFormat(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
