package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

func RandomB64String(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("RandomB64String failed: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func RandomHexString(length int) (string, error) {
	b := make([]byte, (length+1)/2)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("RandomHexString failed: %w", err)
	}
	return hex.EncodeToString(b)[:length], nil
}

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

func RandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("RandomString failed: %w", err)
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}
