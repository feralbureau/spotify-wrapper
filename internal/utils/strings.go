package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
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

func ExtractMappings(jsCode string) (map[string]string, map[string]string) {
	pattern := `\{\d+:"[^"]+"(?:,\d+:"[^"]+")*\}`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(jsCode, -1)

	if len(matches) < 5 {
		return nil, nil
	}

	// need the fourth and fifth matches based on the python implementation
	m1 := parseMap(matches[3])
	m2 := parseMap(matches[4])

	return m1, m2
}

func parseMap(s string) map[string]string {
	res := make(map[string]string)
	s = strings.Trim(s, "{}")
	parts := strings.Split(s, ",")
	for _, p := range parts {
		kv := strings.SplitN(p, ":", 2)
		if len(kv) == 2 {
			k := strings.Trim(kv[0], "\" ")
			v := strings.Trim(kv[1], "\" ")
			res[k] = v
		}
	}
	return res
}

func CombineChunks(nameMap, hashMap map[string]string) []string {
	var combined []string
	for k, v := range nameMap {
		if hv, ok := hashMap[k]; ok {
			combined = append(combined, fmt.Sprintf("%s.%s.js", v, hv))
		}
	}
	return combined
}

func ExtractJSLinks(htmlContent string) []string {
	re := regexp.MustCompile(`<script[^>]*src="([^"]+\.js)"`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)
	var links []string
	for _, m := range matches {
		links = append(links, m[1])
	}
	return links
}
