package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// RandomB64String generates a base64-encoded string produced from cryptographically secure random bytes of the specified length.
func RandomB64String(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// RandomHexString generates a cryptographically secure hexadecimal string of the specified length.
// If length is odd, an extra byte is generated and the hex result is truncated to the requested length.
// The returned string uses lowercase hexadecimal characters.
func RandomHexString(length int) string {
	b := make([]byte, (length+1)/2)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

// ParseJsonString extracts the string value associated with the JSON key s from the input string b.
// It returns the value when a substring matching `"<key>":"<value>"` is found, or an error if the key pattern
// is not present or a closing double quote for the value cannot be found.
func ParseJsonString(b, s string) (string, error) {
	search := fmt.Sprintf(`"%s":"`, s)
	startIndex := strings.Index(b, search)
	if startIndex == -1 {
		return "", fmt.Errorf("substring %s not found", search)
	}

	valueStartIndex := startIndex + len(search)
	valueEndIndex := strings.Index(b[valueStartIndex:], `"`)
	if valueEndIndex == -1 {
		return "", fmt.Errorf("closing double quote not found")
	}

	return b[valueStartIndex : valueStartIndex+valueEndIndex], nil
}

// RandomString returns a cryptographically secure random string of the specified length.
// The result contains only ASCII letters 'a'–'z' and 'A'–'Z'; an empty string is returned if length is zero.
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}
