package spotapi

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/http"
	"github.com/spotapi/spotapi-go/internal/types"
)

func TestNewPublicPlaylist(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	playlist := NewPublicPlaylist("37i9dQZF1DXcBWIGoYBM5M", client, "en")

	if playlist.PlaylistId != "37i9dQZF1DXcBWIGoYBM5M" {
		t.Errorf("Expected playlist ID 37i9dQZF1DXcBWIGoYBM5M, got %s", playlist.PlaylistId)
	}
}

func TestNewPublicPlaylistWithURL(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain ID",
			input:    "37i9dQZF1DXcBWIGoYBM5M",
			expected: "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:     "full URL",
			input:    "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M",
			expected: "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:     "URL with parameters",
			input:    "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M?si=abc123",
			expected: "37i9dQZF1DXcBWIGoYBM5M?si=abc123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			playlist := NewPublicPlaylist(tc.input, client, "en")

			if playlist.PlaylistId != tc.expected {
				t.Errorf("Expected playlist ID %q, got %q", tc.expected, playlist.PlaylistId)
			}

			if playlist.Base == nil {
				t.Error("Base client should not be nil")
			}
		})
	}
}

func TestNewPublicPlaylistLanguages(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	languages := []string{"en", "es", "fr", "de", "ja"}

	for _, lang := range languages {
		playlist := NewPublicPlaylist("test123", client, lang)

		if playlist.Base.Language != lang {
			t.Errorf("Expected language %q, got %q", lang, playlist.Base.Language)
		}
	}
}

func TestNewPrivatePlaylist(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	playlist := NewPrivatePlaylist(login, "test-playlist-id", "en")

	if playlist == nil {
		t.Fatal("NewPrivatePlaylist returned nil")
	}

	if playlist.PlaylistId != "test-playlist-id" {
		t.Errorf("Expected playlist ID 'test-playlist-id', got %q", playlist.PlaylistId)
	}

	if playlist.Login != login {
		t.Error("Login should be set")
	}

	if playlist.Base == nil {
		t.Error("Base should not be nil")
	}
}

func TestNewPrivatePlaylistWithURL(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain ID",
			input:    "abc123",
			expected: "abc123",
		},
		{
			name:     "full URL",
			input:    "https://open.spotify.com/playlist/xyz789",
			expected: "xyz789",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			playlist := NewPrivatePlaylist(login, tc.input, "en")

			if playlist.PlaylistId != tc.expected {
				t.Errorf("Expected playlist ID %q, got %q", tc.expected, playlist.PlaylistId)
			}
		})
	}
}

func TestPublicPlaylistStructure(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	playlist := NewPublicPlaylist("test123", client, "en")

	// Verify structure
	if playlist.Base == nil {
		t.Error("Base should not be nil")
	}

	if playlist.PlaylistId == "" {
		t.Error("PlaylistId should not be empty")
	}

	if playlist.Base.Client == nil {
		t.Error("Base.Client should not be nil")
	}
}

func TestPrivatePlaylistStructure(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	playlist := NewPrivatePlaylist(login, "test123", "es")

	// Verify structure
	if playlist.Base == nil {
		t.Error("Base should not be nil")
	}

	if playlist.Login == nil {
		t.Error("Login should not be nil")
	}

	if playlist.PlaylistId == "" {
		t.Error("PlaylistId should not be empty")
	}

	if playlist.Base.Language != "es" {
		t.Errorf("Expected language 'es', got %q", playlist.Base.Language)
	}
}

func TestPlaylistIDExtraction(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	// Test edge case with multiple "playlist/" in string
	input := "https://example.com/playlist/first/playlist/second"
	playlist := NewPublicPlaylist(input, client, "en")

	// Should extract last occurrence
	if playlist.PlaylistId != "second" {
		t.Errorf("Expected 'second', got %q", playlist.PlaylistId)
	}
}

func TestPrivatePlaylistIDExtraction(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	// Test edge case with multiple "playlist/" in string
	input := "https://example.com/playlist/first/playlist/second"
	playlist := NewPrivatePlaylist(login, input, "en")

	// Should extract last occurrence
	if playlist.PlaylistId != "second" {
		t.Errorf("Expected 'second', got %q", playlist.PlaylistId)
	}
}