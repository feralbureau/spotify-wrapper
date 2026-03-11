package utils

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

func TestRandomB64String(t *testing.T) {
	length := 16
	result := RandomB64String(length)

	// Decode to verify it's valid base64
	decoded, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		t.Errorf("RandomB64String did not produce valid base64: %v", err)
	}

	// Check that decoded length matches input
	if len(decoded) != length {
		t.Errorf("Expected decoded length %d, got %d", length, len(decoded))
	}

	// Test uniqueness - two calls should produce different results
	result2 := RandomB64String(length)
	if result == result2 {
		t.Error("RandomB64String produced identical results on consecutive calls")
	}
}

func TestRandomB64StringDifferentLengths(t *testing.T) {
	testCases := []int{8, 16, 32, 64}

	for _, length := range testCases {
		result := RandomB64String(length)
		decoded, err := base64.StdEncoding.DecodeString(result)
		if err != nil {
			t.Errorf("RandomB64String(%d) did not produce valid base64: %v", length, err)
		}
		if len(decoded) != length {
			t.Errorf("Expected decoded length %d, got %d", length, len(decoded))
		}
	}
}

func TestRandomHexString(t *testing.T) {
	length := 32
	result := RandomHexString(length)

	// Check length
	if len(result) != length {
		t.Errorf("Expected length %d, got %d", length, len(result))
	}

	// Verify it's valid hex
	_, err := hex.DecodeString(result)
	if err != nil {
		t.Errorf("RandomHexString did not produce valid hex: %v", err)
	}

	// Test uniqueness
	result2 := RandomHexString(length)
	if result == result2 {
		t.Error("RandomHexString produced identical results on consecutive calls")
	}
}

func TestRandomHexStringOddLength(t *testing.T) {
	// Test with odd length
	length := 17
	result := RandomHexString(length)

	if len(result) != length {
		t.Errorf("Expected length %d, got %d", length, len(result))
	}

	// With odd length, the hex string itself is odd which can't be decoded to bytes
	// But verify all characters are valid hex characters (0-9, a-f)
	for i, c := range result {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Character at position %d is not a valid hex character: %c", i, c)
		}
	}
}

func TestParseJsonString(t *testing.T) {
	testCases := []struct {
		name     string
		body     string
		search   string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple key-value",
			body:     `{"name":"John Doe","age":"30"}`,
			search:   "name",
			expected: "John Doe",
			wantErr:  false,
		},
		{
			name:     "embedded in larger json",
			body:     `{"user":{"id":"123","username":"testuser","email":"test@example.com"}}`,
			search:   "username",
			expected: "testuser",
			wantErr:  false,
		},
		{
			name:     "key not found",
			body:     `{"name":"John Doe"}`,
			search:   "nonexistent",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty value",
			body:     `{"name":""}`,
			search:   "name",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "value with special characters",
			body:     `{"token":"abc-123_XYZ.456"}`,
			search:   "token",
			expected: "abc-123_XYZ.456",
			wantErr:  false,
		},
		{
			name:     "first occurrence in multiple matches",
			body:     `{"id":"first","data":{"id":"second"}}`,
			search:   "id",
			expected: "first",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseJsonString(tc.body, tc.search)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestParseJsonStringMalformed(t *testing.T) {
	// Test with malformed JSON (missing closing quote)
	body := `{"name":"John Doe`
	_, err := ParseJsonString(body, "name")
	if err == nil {
		t.Error("Expected error for malformed JSON, got none")
	}
}

func TestRandomString(t *testing.T) {
	length := 20
	result := RandomString(length)

	// Check length
	if len(result) != length {
		t.Errorf("Expected length %d, got %d", length, len(result))
	}

	// Check that it only contains letters
	for _, char := range result {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			t.Errorf("RandomString contains non-letter character: %c", char)
		}
	}

	// Test uniqueness
	result2 := RandomString(length)
	if result == result2 {
		t.Error("RandomString produced identical results on consecutive calls")
	}
}

func TestRandomStringDifferentLengths(t *testing.T) {
	testCases := []int{1, 5, 10, 50, 100}

	for _, length := range testCases {
		result := RandomString(length)
		if len(result) != length {
			t.Errorf("Expected length %d, got %d", length, len(result))
		}
	}
}

func TestRandomStringCharacterDistribution(t *testing.T) {
	// Generate a longer string to check for both uppercase and lowercase
	result := RandomString(1000)

	hasLowercase := strings.ContainsAny(result, "abcdefghijklmnopqrstuvwxyz")
	hasUppercase := strings.ContainsAny(result, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	if !hasLowercase {
		t.Error("RandomString did not generate any lowercase characters in 1000 chars")
	}

	if !hasUppercase {
		t.Error("RandomString did not generate any uppercase characters in 1000 chars")
	}
}