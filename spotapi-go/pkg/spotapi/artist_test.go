package spotapi

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/http"
	"github.com/spotapi/spotapi-go/internal/types"
)

func TestNewArtistWithoutLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	artist := NewArtist(nil, client, "en")

	if artist == nil {
		t.Fatal("NewArtist returned nil")
	}

	if artist.Base == nil {
		t.Error("Base should not be nil")
	}

	if artist.login {
		t.Error("Should not be logged in when created without Login")
	}

	if artist.Base.Language != "en" {
		t.Errorf("Expected language 'en', got %q", artist.Base.Language)
	}
}

func TestNewArtistWithLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	artist := NewArtist(login, nil, "es")

	if artist == nil {
		t.Fatal("NewArtist returned nil")
	}

	if !artist.login {
		t.Error("Should be logged in when created with authorized Login")
	}

	if artist.Base.Language != "es" {
		t.Errorf("Expected language 'es', got %q", artist.Base.Language)
	}
}

func TestNewArtistWithUnauthorizedLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: false}

	artist := NewArtist(login, nil, "en")

	if artist.login {
		t.Error("Should not be logged in when Login is not authorized")
	}
}

func TestNewArtistDifferentLanguages(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	languages := []string{"en", "es", "fr", "de", "ja", "pt", "it"}

	for _, lang := range languages {
		artist := NewArtist(nil, client, lang)

		if artist.Base.Language != lang {
			t.Errorf("Expected language %q, got %q", lang, artist.Base.Language)
		}
	}
}

func TestArtistStructure(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	artist := NewArtist(nil, client, "en")

	// Verify structure
	if artist.Base == nil {
		t.Error("Base should not be nil")
	}

	if artist.Base.Client == nil {
		t.Error("Base.Client should not be nil")
	}
}

func TestArtistLoginState(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	testCases := []struct {
		name          string
		login         *Login
		expectedLogin bool
	}{
		{
			name:          "no login",
			login:         nil,
			expectedLogin: false,
		},
		{
			name:          "authorized login",
			login:         &Login{Config: cfg, Authorized: true},
			expectedLogin: true,
		},
		{
			name:          "unauthorized login",
			login:         &Login{Config: cfg, Authorized: false},
			expectedLogin: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			artist := NewArtist(tc.login, client, "en")

			if artist.login != tc.expectedLogin {
				t.Errorf("Expected login state %v, got %v", tc.expectedLogin, artist.login)
			}
		})
	}
}

func TestArtistClientSelection(t *testing.T) {
	client1, _ := http.NewClient(profiles.Chrome_120, "", 3)
	client2, _ := http.NewClient(profiles.Firefox_120, "", 3)

	cfg := &types.Config{Client: client1}
	login := &Login{Config: cfg, Authorized: true}

	// When login is provided, should use login's client
	artist := NewArtist(login, client2, "en")

	// Base.Client should be from login, not the passed client
	if artist.Base.Client != client1 {
		t.Error("Should use client from Login when login is provided")
	}
}

func TestArtistWithoutLoginUsesProvidedClient(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	artist := NewArtist(nil, client, "en")

	if artist.Base.Client != client {
		t.Error("Should use provided client when login is nil")
	}
}

func TestArtistMultipleInstances(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	artist1 := NewArtist(nil, client, "en")
	artist2 := NewArtist(nil, client, "es")

	// Verify they are independent
	if artist1.Base.Language == artist2.Base.Language {
		t.Error("Different artist instances should have different languages")
	}

	// Both should not be logged in
	if artist1.login || artist2.login {
		t.Error("Artist instances without login should not be logged in")
	}
}

func TestArtistFollowWithoutLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	artist := NewArtist(nil, client, "en")

	err := artist.Follow("test-artist-id")

	if err == nil {
		t.Error("Follow should fail when not logged in")
	}

	expectedMsg := "must be logged in"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestArtistUnfollowWithoutLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	artist := NewArtist(nil, client, "en")

	err := artist.Unfollow("test-artist-id")

	if err == nil {
		t.Error("Unfollow should fail when not logged in")
	}

	expectedMsg := "must be logged in"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestArtistIDExtraction(t *testing.T) {
	// This tests the logic in doFollow which strips "artist:" prefix
	// We can't fully test without mocking, but we can test the struct creation

	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	artist := NewArtist(nil, client, "en")

	if artist == nil {
		t.Fatal("Artist creation failed")
	}
}

func TestNewArtistNilChecks(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	// Should not panic with nil login
	artist := NewArtist(nil, client, "en")
	if artist == nil {
		t.Error("Should handle nil login gracefully")
	}
}

func TestArtistBaseClientConfiguration(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	artist := NewArtist(login, nil, "fr")

	// Verify BaseClient is properly configured
	if artist.Base == nil {
		t.Fatal("Base should not be nil")
	}

	if artist.Base.Language != "fr" {
		t.Errorf("Expected language 'fr', got %q", artist.Base.Language)
	}

	if artist.Base.Client == nil {
		t.Error("Base.Client should not be nil")
	}
}