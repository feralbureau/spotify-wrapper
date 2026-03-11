package spotapi

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/http"
)

func TestNewPublicAlbum(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	testCases := []struct {
		name        string
		input       string
		expectedId  string
		expectedUrl string
	}{
		{
			name:        "plain album ID",
			input:       "3IBcauSj5M2A6lTeffJzdv",
			expectedId:  "3IBcauSj5M2A6lTeffJzdv",
			expectedUrl: "https://open.spotify.com/album/3IBcauSj5M2A6lTeffJzdv",
		},
		{
			name:        "full album URL",
			input:       "https://open.spotify.com/album/6DEjYFkNZh67HP7R9PSZvv",
			expectedId:  "6DEjYFkNZh67HP7R9PSZvv",
			expectedUrl: "https://open.spotify.com/album/6DEjYFkNZh67HP7R9PSZvv",
		},
		{
			name:        "URL with query params",
			input:       "https://open.spotify.com/album/1A2GTWGtFfWp7KSQTwWOyo?si=abc123",
			expectedId:  "1A2GTWGtFfWp7KSQTwWOyo?si=abc123",
			expectedUrl: "https://open.spotify.com/album/1A2GTWGtFfWp7KSQTwWOyo?si=abc123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			album := NewPublicAlbum(tc.input, client, "en")

			if album == nil {
				t.Fatal("NewPublicAlbum returned nil")
			}

			if album.AlbumId != tc.expectedId {
				t.Errorf("Expected album ID %q, got %q", tc.expectedId, album.AlbumId)
			}

			if album.AlbumLink != tc.expectedUrl {
				t.Errorf("Expected album link %q, got %q", tc.expectedUrl, album.AlbumLink)
			}

			if album.Base == nil {
				t.Error("Base client should not be nil")
			}
		})
	}
}

func TestNewPublicAlbumLanguages(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	languages := []string{"en", "es", "fr", "de", "ja", "pt"}

	for _, lang := range languages {
		album := NewPublicAlbum("test123", client, lang)

		if album.Base.Language != lang {
			t.Errorf("Expected language %q, got %q", lang, album.Base.Language)
		}
	}
}

func TestPublicAlbumStructure(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	album := NewPublicAlbum("test-album-id", client, "en")

	// Verify all fields are properly set
	if album.Base == nil {
		t.Error("Base should not be nil")
	}

	if album.AlbumId == "" {
		t.Error("AlbumId should not be empty")
	}

	if album.AlbumLink == "" {
		t.Error("AlbumLink should not be empty")
	}

	if album.Base.Client == nil {
		t.Error("Base.Client should not be nil")
	}
}

func TestPublicAlbumIDExtraction(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	// Test with multiple "album/" in string
	input := "https://example.com/album/first/album/second"
	album := NewPublicAlbum(input, client, "en")

	// Should extract the last occurrence
	if album.AlbumId != "second" {
		t.Errorf("Expected 'second', got %q", album.AlbumId)
	}
}

func TestPublicAlbumLinkGeneration(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	testCases := []struct {
		name        string
		input       string
		expectedUrl string
	}{
		{
			name:        "simple ID",
			input:       "abc123",
			expectedUrl: "https://open.spotify.com/album/abc123",
		},
		{
			name:        "ID with special characters",
			input:       "abc-123_XYZ",
			expectedUrl: "https://open.spotify.com/album/abc-123_XYZ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			album := NewPublicAlbum(tc.input, client, "en")

			if album.AlbumLink != tc.expectedUrl {
				t.Errorf("Expected album link %q, got %q", tc.expectedUrl, album.AlbumLink)
			}
		})
	}
}

func TestPublicAlbumWithEmptyID(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	album := NewPublicAlbum("", client, "en")

	if album.AlbumId != "" {
		t.Error("Empty input should result in empty AlbumId")
	}

	if album.AlbumLink != "https://open.spotify.com/album/" {
		t.Errorf("Expected base URL, got %q", album.AlbumLink)
	}
}

func TestPublicAlbumClientAssignment(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	album := NewPublicAlbum("test123", client, "en")

	if album.Base.Client != client {
		t.Error("Base.Client should reference the provided client")
	}
}

func TestPublicAlbumMultipleInstances(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	album1 := NewPublicAlbum("album1", client, "en")
	album2 := NewPublicAlbum("album2", client, "es")

	// Verify they are independent
	if album1.AlbumId == album2.AlbumId {
		t.Error("Different albums should have different IDs")
	}

	if album1.Base.Language == album2.Base.Language {
		t.Error("Different albums should have different languages")
	}
}

func TestPublicAlbumURLParsing(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	// Test various URL formats
	urls := []struct {
		input      string
		expectedId string
	}{
		{
			"https://open.spotify.com/album/123abc",
			"123abc",
		},
		{
			"http://open.spotify.com/album/456def",
			"456def",
		},
		{
			"open.spotify.com/album/789ghi",
			"789ghi",
		},
		{
			"spotify.com/album/xyz",
			"xyz",
		},
	}

	for _, u := range urls {
		album := NewPublicAlbum(u.input, client, "en")
		if album.AlbumId != u.expectedId {
			t.Errorf("Input %q: expected ID %q, got %q", u.input, u.expectedId, album.AlbumId)
		}
	}
}