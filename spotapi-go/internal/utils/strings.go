package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

func RandomB64String(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func RandomHexString(length int) string {
	b := make([]byte, (length+1)/2)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
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

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}
