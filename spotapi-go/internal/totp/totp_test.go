package totp

import (
	"encoding/base32"
	"regexp"
	"testing"
)

func TestGenerateTOTP(t *testing.T) {
	totp, version := GenerateTOTP()

	// Check TOTP format - should be 6 digits
	if len(totp) != 6 {
		t.Errorf("Expected TOTP length 6, got %d", len(totp))
	}

	// Check that it's all digits
	matched, _ := regexp.MatchString("^[0-9]{6}$", totp)
	if !matched {
		t.Errorf("TOTP should be 6 digits, got %q", totp)
	}

	// Check version is reasonable
	if version < 0 {
		t.Errorf("Version should be non-negative, got %d", version)
	}
}

func TestGenerateTOTPUniqueness(t *testing.T) {
	// Generate multiple TOTPs in quick succession
	// They should be the same (same time window) but we test the function works consistently
	totp1, version1 := GenerateTOTP()
	totp2, version2 := GenerateTOTP()

	// Within the same 30-second window, TOTPs should be identical
	if totp1 != totp2 {
		t.Logf("TOTPs differ (might be due to timing boundary): %s vs %s", totp1, totp2)
	}

	// Versions should be the same
	if version1 != version2 {
		t.Errorf("Expected same version, got %d and %d", version1, version2)
	}
}

func TestGenerateTOTPFromSecret(t *testing.T) {
	// Test with a known secret
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("test secret key"))

	totp := generateTOTPFromSecret(secret)

	// Check format
	if len(totp) != 6 {
		t.Errorf("Expected TOTP length 6, got %d", len(totp))
	}

	// Check that it's all digits
	matched, _ := regexp.MatchString("^[0-9]{6}$", totp)
	if !matched {
		t.Errorf("TOTP should be 6 digits, got %q", totp)
	}

	// Generate again with same secret - should be same in same time window
	totp2 := generateTOTPFromSecret(secret)
	if totp != totp2 {
		t.Errorf("TOTP with same secret should be identical within time window: %s vs %s", totp, totp2)
	}
}

func TestGenerateTOTPFromSecretDifferentSecrets(t *testing.T) {
	secret1 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("secret one"))
	secret2 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("secret two"))

	totp1 := generateTOTPFromSecret(secret1)
	totp2 := generateTOTPFromSecret(secret2)

	// Different secrets should (very likely) produce different TOTPs
	if totp1 == totp2 {
		t.Logf("Warning: Different secrets produced same TOTP (collision): %s", totp1)
	}
}

func TestGetLatestTotpSecret(t *testing.T) {
	version, secret := GetLatestTotpSecret()

	// Check version is reasonable
	if version < 0 {
		t.Errorf("Version should be non-negative, got %d", version)
	}

	// Check secret is not empty
	if len(secret) == 0 {
		t.Error("Secret should not be empty")
	}

	// If network fails, should return fallback values
	if version == fallbackVer && len(secret) == len(fallbackSecret) {
		t.Logf("Using fallback secret (network may have failed)")
	}
}

func TestGetLatestTotpSecretFallback(t *testing.T) {
	// The function should handle network failures gracefully
	// This test just verifies the fallback values are defined
	if fallbackVer < 0 {
		t.Errorf("Fallback version should be non-negative, got %d", fallbackVer)
	}

	if len(fallbackSecret) == 0 {
		t.Error("Fallback secret should not be empty")
	}
}

func TestTOTPConsistency(t *testing.T) {
	// Test that multiple calls in rapid succession produce valid results
	for i := 0; i < 10; i++ {
		totp, version := GenerateTOTP()

		if len(totp) != 6 {
			t.Errorf("Iteration %d: Expected TOTP length 6, got %d", i, len(totp))
		}

		matched, _ := regexp.MatchString("^[0-9]{6}$", totp)
		if !matched {
			t.Errorf("Iteration %d: TOTP should be 6 digits, got %q", i, totp)
		}

		if version < 0 {
			t.Errorf("Iteration %d: Version should be non-negative, got %d", i, version)
		}
	}
}

func TestGenerateTOTPWithEmptySecret(t *testing.T) {
	// Test edge case with empty secret
	secret := ""
	totp := generateTOTPFromSecret(secret)

	// Should still produce a 6-digit code (even if not useful)
	if len(totp) != 6 {
		t.Errorf("Expected TOTP length 6, got %d", len(totp))
	}

	matched, _ := regexp.MatchString("^[0-9]{6}$", totp)
	if !matched {
		t.Errorf("TOTP should be 6 digits, got %q", totp)
	}
}

func TestTOTPFormat(t *testing.T) {
	// Test that TOTP is properly zero-padded
	for i := 0; i < 20; i++ {
		totp, _ := GenerateTOTP()

		// All TOTPs should be exactly 6 characters
		if len(totp) != 6 {
			t.Errorf("TOTP length should always be 6, got %d: %q", len(totp), totp)
		}

		// Should start with digits (including potential leading zeros)
		for _, char := range totp {
			if char < '0' || char > '9' {
				t.Errorf("TOTP contains non-digit character: %c in %q", char, totp)
			}
		}
	}
}