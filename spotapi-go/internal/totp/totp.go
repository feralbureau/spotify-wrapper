package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	fallbackSecret = []byte{70, 60, 33, 57, 92, 120, 90, 33, 32, 62, 62, 55, 126, 93, 66, 35, 108, 68}
	fallbackVer    = 18
)

// GetLatestTotpSecret retrieves the latest available TOTP secret version and its secret bytes from the remote secret dictionary.
// If fetching, reading, or parsing the remote data fails or no valid version is found, it returns the package fallback version and secret.
// The first return value is the numeric version; the second is the secret as a byte slice.
func GetLatestTotpSecret() (int, []byte) {
	resp, err := http.Get("https://code.thetadev.de/ThetaDev/spotify-secrets/raw/branch/main/secrets/secretDict.json")
	if err != nil {
		return fallbackVer, fallbackSecret
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fallbackVer, fallbackSecret
	}

	var secrets map[string][]int
	if err := json.Unmarshal(body, &secrets); err != nil {
		return fallbackVer, fallbackSecret
	}

	maxVer := -1
	for k := range secrets {
		v, _ := strconv.Atoi(k)
		if v > maxVer {
			maxVer = v
		}
	}

	if maxVer == -1 {
		return fallbackVer, fallbackSecret
	}

	verStr := strconv.Itoa(maxVer)
	secretInts := secrets[verStr]
	secretBytes := make([]byte, len(secretInts))
	for i, v := range secretInts {
		secretBytes[i] = byte(v)
	}

	return maxVer, secretBytes
}

// GenerateTOTP produces a 6-digit TOTP code derived from the latest available secret and returns it along with that secret's version.
// It obtains the most recent secret, transforms and encodes it into a base32 key, and computes the current TOTP using that key.
// The first return value is the 6-digit TOTP string; the second is the numeric version of the secret used.
func GenerateTOTP() (string, int) {
	version, secretBytes := GetLatestTotpSecret()
	transformed := make([]byte, len(secretBytes))
	for i, b := range secretBytes {
		transformed[i] = b ^ byte((i%33)+9)
	}

	joined := ""
	for _, b := range transformed {
		joined += fmt.Sprintf("%d", b)
	}

	hexStr := hex.EncodeToString([]byte(joined))
	decodedHex, _ := hex.DecodeString(hexStr)
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(decodedHex)

	return generateTOTPFromSecret(secret), version
}

// generateTOTPFromSecret computes a 6-digit TOTP code from a base32-encoded secret using the current 30-second interval.
// It decodes the secret, computes HMAC-SHA1 over the interval counter, applies dynamic truncation, and returns a zero-padded 6-digit decimal string.
func generateTOTPFromSecret(secret string) string {
	key, _ := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	interval := time.Now().Unix() / 30
	counter := make([]byte, 8)
	binary.BigEndian.PutUint64(counter, uint64(interval))

	h := hmac.New(sha1.New, key)
	h.Write(counter)
	sum := h.Sum(nil)

	offset := sum[len(sum)-1] & 0xf
	code := int32(sum[offset]&0x7f)<<24 |
		int32(sum[offset+1])<<16 |
		int32(sum[offset+2])<<8 |
		int32(sum[offset+3])

	return fmt.Sprintf("%06d", code%1000000)
}
