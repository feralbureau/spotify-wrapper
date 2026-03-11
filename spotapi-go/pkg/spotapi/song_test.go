package spotapi

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/http"
	"github.com/spotapi/spotapi-go/internal/types"
)

func TestNewSongWithoutPlaylist(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	song := NewSong(nil, client, "en")

	if song == nil {
		t.Fatal("NewSong returned nil")
	}

	if song.Playlist != nil {
		t.Error("Playlist should be nil when not provided")
	}

	if song.Base == nil {
		t.Error("Base should not be nil")
	}

	if song.Base.Language != "en" {
		t.Errorf("Expected language 'en', got %q", song.Base.Language)
	}
}

func TestNewSongWithPlaylist(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}
	playlist := &PrivatePlaylist{
		Login:      login,
		PlaylistId: "test-playlist",
	}

	song := NewSong(playlist, nil, "es")

	if song == nil {
		t.Fatal("NewSong returned nil")
	}

	if song.Playlist != playlist {
		t.Error("Playlist should be set")
	}

	if song.Base == nil {
		t.Error("Base should not be nil")
	}

	if song.Base.Language != "es" {
		t.Errorf("Expected language 'es', got %q", song.Base.Language)
	}
}

func TestNewSongDifferentLanguages(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	languages := []string{"en", "es", "fr", "de", "ja", "pt", "it", "ko", "zh"}

	for _, lang := range languages {
		song := NewSong(nil, client, lang)

		if song.Base.Language != lang {
			t.Errorf("Expected language %q, got %q", lang, song.Base.Language)
		}
	}
}

func TestSongStructure(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	song := NewSong(nil, client, "en")

	// Verify structure
	if song.Base == nil {
		t.Error("Base should not be nil")
	}

	if song.Base.Client == nil {
		t.Error("Base.Client should not be nil")
	}
}

func TestSongClientSelection(t *testing.T) {
	client1, _ := http.NewClient(profiles.Chrome_120, "", 3)
	client2, _ := http.NewClient(profiles.Firefox_120, "", 3)

	cfg := &types.Config{Client: client1}
	login := &Login{Config: cfg, Authorized: true}
	playlist := &PrivatePlaylist{
		Login:      login,
		PlaylistId: "test",
	}

	// When playlist is provided, should use playlist's client
	song := NewSong(playlist, client2, "en")

	// Base.Client should be from playlist's login, not the passed client
	if song.Base.Client != client1 {
		t.Error("Should use client from Playlist when playlist is provided")
	}
}

func TestSongWithoutPlaylistUsesProvidedClient(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	song := NewSong(nil, client, "en")

	if song.Base.Client != client {
		t.Error("Should use provided client when playlist is nil")
	}
}

func TestAddSongsToPlaylistWithoutPlaylist(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	song := NewSong(nil, client, "en")

	err := song.AddSongsToPlaylist([]string{"track1", "track2"})

	if err == nil {
		t.Error("AddSongsToPlaylist should fail when playlist is not set")
	}

	expectedMsg := "playlist not set"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestAddSongsToPlaylistWithEmptyPlaylistId(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}
	playlist := &PrivatePlaylist{
		Login:      login,
		PlaylistId: "",
	}

	song := NewSong(playlist, nil, "en")

	err := song.AddSongsToPlaylist([]string{"track1"})

	if err == nil {
		t.Error("AddSongsToPlaylist should fail when playlist ID is empty")
	}
}

func TestSongMultipleInstances(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	song1 := NewSong(nil, client, "en")
	song2 := NewSong(nil, client, "es")

	// Verify they are independent
	if song1.Base.Language == song2.Base.Language {
		t.Error("Different song instances should have different languages")
	}
}

func TestSongBaseClientConfiguration(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}
	playlist := &PrivatePlaylist{
		Login:      login,
		PlaylistId: "test",
	}

	song := NewSong(playlist, nil, "fr")

	// Verify BaseClient is properly configured
	if song.Base == nil {
		t.Fatal("Base should not be nil")
	}

	if song.Base.Language != "fr" {
		t.Errorf("Expected language 'fr', got %q", song.Base.Language)
	}

	if song.Base.Client == nil {
		t.Error("Base.Client should not be nil")
	}
}

func TestAddSongsToPlaylistEmptyList(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}
	playlist := &PrivatePlaylist{
		Login:      login,
		PlaylistId: "test-playlist",
	}

	song := NewSong(playlist, nil, "en")

	// Adding empty list - function will still attempt the request
	// In real scenario this would make API call, here we just test structure
	err := song.AddSongsToPlaylist([]string{})

	// Will fail because we can't make real API calls in unit test
	if err == nil {
		t.Log("AddSongsToPlaylist succeeded (may have network access)")
	} else {
		t.Logf("Expected failure in unit test without network: %v", err)
	}
}

func TestSongNilChecks(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)

	// Should not panic with nil playlist
	song := NewSong(nil, client, "en")
	if song == nil {
		t.Error("Should handle nil playlist gracefully")
	}

	// Should not panic with nil client when playlist is provided
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}
	playlist := &PrivatePlaylist{
		Login:      login,
		PlaylistId: "test",
	}

	song2 := NewSong(playlist, nil, "en")
	if song2 == nil {
		t.Error("Should handle nil client gracefully when playlist is provided")
	}
}

func TestSongPlaylistAssignment(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}
	playlist := &PrivatePlaylist{
		Login:      login,
		PlaylistId: "test123",
	}

	song := NewSong(playlist, nil, "en")

	if song.Playlist != playlist {
		t.Error("Playlist should be properly assigned")
	}

	if song.Playlist.PlaylistId != "test123" {
		t.Error("Playlist ID should be accessible through Song.Playlist")
	}
}