package querytracker

import (
	"testing"
)

func TestDetectIPType(t *testing.T) {
	tests := []struct {
		ip       string
		expected string
	}{
		{"192.168.1.1", "IPv4"},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "IPv6"},
		{"invalid_ip", ""},
	}

	for _, test := range tests {
		result, err := detect_ip_type(test.ip)
		if err != nil && test.expected != "" {
			t.Errorf("Expected no error for IP %s, but got %v", test.ip, err)
		}
		if result != test.expected {
			t.Errorf("Expected %s for IP %s, but got %s", test.expected, test.ip, result)
		}
	}
}

func TestIsIPInRange(t *testing.T) {
	tests := []struct {
		ip       string
		cidr     string
		expected bool
	}{
		{"192.168.1.1", "192.168.1.0/24", true},
		{"192.168.1.1", "192.168.2.0/24", false},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:db8::/32", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:db9::/32", false},
	}

	for _, test := range tests {
		result, err := is_ip_in_range(test.ip, test.cidr)
		if err != nil {
			t.Errorf("Expected no error for IP %s and CIDR %s, but got %v", test.ip, test.cidr, err)
		}
		if result != test.expected {
			t.Errorf("Expected %v for IP %s and CIDR %s, but got %v", test.expected, test.ip, test.cidr, result)
		}
	}
}