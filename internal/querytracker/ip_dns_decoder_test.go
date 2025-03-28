package querytracker

import (
	"testing"
)

func TestDecodeDNSIP(t *testing.T) {
	tests := []struct {
		name       string
		encodedStr string
		expected   string
		expectErr  bool
	}{
		{
			name:       "Valid IPv4",
			encodedStr: "MTkyLjE2OC4xLjE=", 
			expected:   "192.168.1.1",
			expectErr:  false,
		},
		{
			name:       "Valid IPv6",
			encodedStr: "MjAwMToxOjE6MToxOjE6MTox",
			expected:   "2001:1:1:1:1:1:1:1",
			expectErr:  false,
		},
		{
			name:       "Invalid data",
			encodedStr: "AAAA",
			expected:   "",
			expectErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := decodeDNSIP(test.encodedStr)
			if (err != nil) != test.expectErr {
				t.Errorf("Expected error: %v for encodedStr %s, but got %v", test.expectErr, test.encodedStr, err)
			}
			if result != test.expected {
				t.Errorf("Expected %s for encodedStr %s, but got %s", test.expected, test.encodedStr, result)
			}
		})
	}
}
