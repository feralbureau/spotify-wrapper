package http

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
)

func TestNewBaseClient(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	if baseClient == nil {
		t.Fatal("NewBaseClient returned nil")
	}

	if baseClient.Client != client {
		t.Error("BaseClient.Client should reference the provided client")
	}

	if baseClient.Language != "en" {
		t.Errorf("Expected language 'en', got %q", baseClient.Language)
	}

	if baseClient.Client.Authenticate == nil {
		t.Error("Client.Authenticate should be set by NewBaseClient")
	}
}

func TestNewBaseClientDifferentLanguages(t *testing.T) {
	languages := []string{"en", "es", "fr", "de", "ja"}

	for _, lang := range languages {
		client, err := NewClient(profiles.Chrome_120, "", 3)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		baseClient := NewBaseClient(client, lang)

		if baseClient.Language != lang {
			t.Errorf("Expected language %q, got %q", lang, baseClient.Language)
		}
	}
}

func TestAuthRuleInitialState(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	// Initially, tokens should be empty
	if baseClient.AccessToken != "" {
		t.Error("AccessToken should be empty initially")
	}

	if baseClient.ClientToken != "" {
		t.Error("ClientToken should be empty initially")
	}

	if baseClient.ClientId != "" {
		t.Error("ClientId should be empty initially")
	}

	if baseClient.DeviceId != "" {
		t.Error("DeviceId should be empty initially")
	}
}

func TestAuthRuleWithNilHeaders(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	// Manually set tokens to avoid network calls
	baseClient.AccessToken = "test-access-token"
	baseClient.ClientToken = "test-client-token"
	baseClient.ClientVersion = "1.0.0"

	// Call AuthRule with nil headers
	headers, err := baseClient.AuthRule(nil)

	if err != nil {
		t.Errorf("AuthRule should not error with valid tokens: %v", err)
	}

	if headers == nil {
		t.Fatal("AuthRule should return non-nil headers")
	}

	// Check headers
	if headers["Authorization"] != "Bearer test-access-token" {
		t.Errorf("Expected Authorization header, got %q", headers["Authorization"])
	}

	if headers["Client-Token"] != "test-client-token" {
		t.Errorf("Expected Client-Token header, got %q", headers["Client-Token"])
	}

	if headers["Spotify-App-Version"] != "1.0.0" {
		t.Errorf("Expected Spotify-App-Version header, got %q", headers["Spotify-App-Version"])
	}

	if headers["Accept-Language"] != "en" {
		t.Errorf("Expected Accept-Language header 'en', got %q", headers["Accept-Language"])
	}
}

func TestAuthRuleWithExistingHeaders(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	// Manually set tokens
	baseClient.AccessToken = "test-access-token"
	baseClient.ClientToken = "test-client-token"
	baseClient.ClientVersion = "1.0.0"

	// Existing headers
	existingHeaders := map[string]string{
		"User-Agent":   "Test/1.0",
		"Content-Type": "application/json",
	}

	headers, err := baseClient.AuthRule(existingHeaders)

	if err != nil {
		t.Errorf("AuthRule failed: %v", err)
	}

	// Check existing headers are preserved
	if headers["User-Agent"] != "Test/1.0" {
		t.Error("Existing User-Agent header should be preserved")
	}

	if headers["Content-Type"] != "application/json" {
		t.Error("Existing Content-Type header should be preserved")
	}

	// Check new headers are added
	if headers["Authorization"] == "" {
		t.Error("Authorization header should be added")
	}
}

func TestPartHashExtraction(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	// Mock RawHashes
	baseClient.RawHashes = `{"fetchPlaylist","query","abc123hash456","getAlbum","mutation","def789hash012"}`

	// Test query extraction
	hash := baseClient.PartHash("fetchPlaylist")
	if hash != "abc123hash456" {
		t.Errorf("Expected hash 'abc123hash456', got %q", hash)
	}

	// Test mutation extraction
	hash = baseClient.PartHash("getAlbum")
	if hash != "def789hash012" {
		t.Errorf("Expected hash 'def789hash012', got %q", hash)
	}

	// Test non-existent hash
	hash = baseClient.PartHash("nonexistent")
	if hash != "" {
		t.Errorf("Expected empty string for non-existent hash, got %q", hash)
	}
}

func TestPartHashWithEmptyRawHashes(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	// RawHashes is empty initially
	hash := baseClient.PartHash("someOperation")

	// Should return empty string when RawHashes is empty
	if hash != "" {
		t.Errorf("Expected empty string when RawHashes is empty, got %q", hash)
	}
}

func TestBaseClientStructFields(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	// Test that all fields are accessible
	_ = baseClient.Client
	_ = baseClient.JsPack
	_ = baseClient.ClientVersion
	_ = baseClient.AccessToken
	_ = baseClient.ClientToken
	_ = baseClient.ClientId
	_ = baseClient.DeviceId
	_ = baseClient.RawHashes
	_ = baseClient.Language
	_ = baseClient.ServerCfg
}

func TestAuthRuleWithoutTokens(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	// Without setting tokens, AuthRule should try to fetch them
	// This will fail in unit test without real network, but we test the flow
	_, err = baseClient.AuthRule(nil)

	// We expect an error since we can't actually fetch tokens in a unit test
	if err == nil {
		t.Log("AuthRule succeeded (may have network access or cached state)")
	} else {
		t.Logf("Expected failure in unit test without network: %v", err)
	}
}

func TestPartHashQueryFormat(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	// Test with proper query format
	baseClient.RawHashes = `"searchDesktop","query","queryhash123"`

	hash := baseClient.PartHash("searchDesktop")
	if hash != "queryhash123" {
		t.Errorf("Expected 'queryhash123', got %q", hash)
	}
}

func TestPartHashMutationFormat(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "en")

	// Test with proper mutation format
	baseClient.RawHashes = `"addToLibrary","mutation","mutationhash456"`

	hash := baseClient.PartHash("addToLibrary")
	if hash != "mutationhash456" {
		t.Errorf("Expected 'mutationhash456', got %q", hash)
	}
}

func TestAuthRuleHeadersOrder(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	baseClient := NewBaseClient(client, "fr")

	// Set required tokens
	baseClient.AccessToken = "token1"
	baseClient.ClientToken = "token2"
	baseClient.ClientVersion = "2.0.0"

	headers, err := baseClient.AuthRule(nil)
	if err != nil {
		t.Fatalf("AuthRule failed: %v", err)
	}

	// Verify all required headers are present
	requiredHeaders := []string{"Authorization", "Client-Token", "Spotify-App-Version", "Accept-Language"}
	for _, header := range requiredHeaders {
		if _, ok := headers[header]; !ok {
			t.Errorf("Required header %q is missing", header)
		}
	}
}