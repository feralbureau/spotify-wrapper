package spotapi

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/http"
)

func TestNewPodcast(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	testCases := []struct {
		name        string
		input       string
		expectedId  string
		expectedUrl string
	}{
		{
			name:        "plain podcast ID",
			input:       "4rOoJ6Egrf8K2IrywzwOMk",
			expectedId:  "4rOoJ6Egrf8K2IrywzwOMk",
			expectedUrl: "https://open.spotify.com/show/4rOoJ6Egrf8K2IrywzwOMk",
		},
		{
			name:        "full podcast URL",
			input:       "https://open.spotify.com/show/2MAi0BvDc6GTFvKFPXnkCL",
			expectedId:  "2MAi0BvDc6GTFvKFPXnkCL",
			expectedUrl: "https://open.spotify.com/show/2MAi0BvDc6GTFvKFPXnkCL",
		},
		{
			name:        "URL with query params",
			input:       "https://open.spotify.com/show/0ofXAdFIQQRsCYj9754UFx?si=xyz789",
			expectedId:  "0ofXAdFIQQRsCYj9754UFx?si=xyz789",
			expectedUrl: "https://open.spotify.com/show/0ofXAdFIQQRsCYj9754UFx?si=xyz789",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			podcast := NewPodcast(tc.input, client, "en")

			if podcast == nil {
				t.Fatal("NewPodcast returned nil")
			}

			if podcast.PodcastId != tc.expectedId {
				t.Errorf("Expected podcast ID %q, got %q", tc.expectedId, podcast.PodcastId)
			}

			if podcast.PodcastLink != tc.expectedUrl {
				t.Errorf("Expected podcast link %q, got %q", tc.expectedUrl, podcast.PodcastLink)
			}

			if podcast.Base == nil {
				t.Error("Base client should not be nil")
			}
		})
	}
}

func TestNewPodcastLanguages(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	languages := []string{"en", "es", "fr", "de", "ja", "pt", "it", "ko"}

	for _, lang := range languages {
		podcast := NewPodcast("test123", client, lang)

		if podcast.Base.Language != lang {
			t.Errorf("Expected language %q, got %q", lang, podcast.Base.Language)
		}
	}
}

func TestPodcastStructure(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	podcast := NewPodcast("test-podcast-id", client, "en")

	// Verify all fields are properly set
	if podcast.Base == nil {
		t.Error("Base should not be nil")
	}

	if podcast.PodcastId == "" {
		t.Error("PodcastId should not be empty")
	}

	if podcast.PodcastLink == "" {
		t.Error("PodcastLink should not be empty")
	}

	if podcast.Base.Client == nil {
		t.Error("Base.Client should not be nil")
	}
}

func TestPodcastIDExtraction(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	// Test with multiple "show/" in string
	input := "https://example.com/show/first/show/second"
	podcast := NewPodcast(input, client, "en")

	// Should extract the last occurrence
	if podcast.PodcastId != "second" {
		t.Errorf("Expected 'second', got %q", podcast.PodcastId)
	}
}

func TestPodcastLinkGeneration(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	testCases := []struct {
		name        string
		input       string
		expectedUrl string
	}{
		{
			name:        "simple ID",
			input:       "abc123",
			expectedUrl: "https://open.spotify.com/show/abc123",
		},
		{
			name:        "ID with special characters",
			input:       "abc-123_XYZ",
			expectedUrl: "https://open.spotify.com/show/abc-123_XYZ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			podcast := NewPodcast(tc.input, client, "en")

			if podcast.PodcastLink != tc.expectedUrl {
				t.Errorf("Expected podcast link %q, got %q", tc.expectedUrl, podcast.PodcastLink)
			}
		})
	}
}

func TestPodcastWithEmptyID(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	podcast := NewPodcast("", client, "en")

	if podcast.PodcastId != "" {
		t.Error("Empty input should result in empty PodcastId")
	}

	if podcast.PodcastLink != "https://open.spotify.com/show/" {
		t.Errorf("Expected base URL, got %q", podcast.PodcastLink)
	}
}

func TestPodcastClientAssignment(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	podcast := NewPodcast("test123", client, "en")

	if podcast.Base.Client != client {
		t.Error("Base.Client should reference the provided client")
	}
}

func TestPodcastMultipleInstances(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	podcast1 := NewPodcast("podcast1", client, "en")
	podcast2 := NewPodcast("podcast2", client, "es")

	// Verify they are independent
	if podcast1.PodcastId == podcast2.PodcastId {
		t.Error("Different podcasts should have different IDs")
	}

	if podcast1.Base.Language == podcast2.Base.Language {
		t.Error("Different podcasts should have different languages")
	}
}

func TestPodcastURLParsing(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	// Test various URL formats
	urls := []struct {
		input      string
		expectedId string
	}{
		{
			"https://open.spotify.com/show/123abc",
			"123abc",
		},
		{
			"http://open.spotify.com/show/456def",
			"456def",
		},
		{
			"open.spotify.com/show/789ghi",
			"789ghi",
		},
		{
			"spotify.com/show/xyz",
			"xyz",
		},
	}

	for _, u := range urls {
		podcast := NewPodcast(u.input, client, "en")
		if podcast.PodcastId != u.expectedId {
			t.Errorf("Input %q: expected ID %q, got %q", u.input, u.expectedId, podcast.PodcastId)
		}
	}
}

func TestPodcastShowKeyword(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	// Verify "show/" is used instead of "podcast/"
	input := "https://open.spotify.com/show/test123"
	podcast := NewPodcast(input, client, "en")

	if podcast.PodcastId != "test123" {
		t.Errorf("Expected 'test123', got %q", podcast.PodcastId)
	}

	// Verify link uses "show/"
	if podcast.PodcastLink != "https://open.spotify.com/show/test123" {
		t.Errorf("PodcastLink should use 'show/', got %q", podcast.PodcastLink)
	}
}

func TestPodcastBaseClientConfiguration(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	podcast := NewPodcast("test123", client, "fr")

	// Verify BaseClient is properly configured
	if podcast.Base == nil {
		t.Fatal("Base should not be nil")
	}

	if podcast.Base.Language != "fr" {
		t.Errorf("Expected language 'fr', got %q", podcast.Base.Language)
	}

	if podcast.Base.Client == nil {
		t.Error("Base.Client should not be nil")
	}
}