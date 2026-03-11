package errors

import (
	"strings"
	"testing"
)

func TestSpotError(t *testing.T) {
	t.Run("with both message and err", func(t *testing.T) {
		err := &SpotError{
			Message: "Test error",
			Err:     "underlying error",
		}
		expected := "Test error: underlying error"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("with only message", func(t *testing.T) {
		err := &SpotError{
			Message: "Test error",
			Err:     "",
		}
		expected := "Test error"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("empty strings", func(t *testing.T) {
		err := &SpotError{
			Message: "",
			Err:     "",
		}
		if err.Error() != "" {
			t.Errorf("Expected empty string, got %q", err.Error())
		}
	})
}

func TestNewSpotError(t *testing.T) {
	msg := "test message"
	errStr := "test error"

	err := NewSpotError(msg, errStr)

	if err == nil {
		t.Fatal("NewSpotError returned nil")
	}

	if err.Message != msg {
		t.Errorf("Expected message %q, got %q", msg, err.Message)
	}

	if err.Err != errStr {
		t.Errorf("Expected err %q, got %q", errStr, err.Err)
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestLoginError(t *testing.T) {
	msg := "login failed"
	errStr := "invalid credentials"

	err := NewLoginError(msg, errStr)

	if err == nil {
		t.Fatal("NewLoginError returned nil")
	}

	if err.SpotError == nil {
		t.Fatal("LoginError.SpotError is nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}

	// Verify it's a LoginError type
	var _ *LoginError = err
}

func TestRequestError(t *testing.T) {
	msg := "request failed"
	errStr := "network timeout"

	err := NewRequestError(msg, errStr)

	if err == nil {
		t.Fatal("NewRequestError returned nil")
	}

	if err.SpotError == nil {
		t.Fatal("RequestError.SpotError is nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestBaseClientError(t *testing.T) {
	msg := "client error"
	errStr := "initialization failed"

	err := NewBaseClientError(msg, errStr)

	if err == nil {
		t.Fatal("NewBaseClientError returned nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestPlaylistError(t *testing.T) {
	msg := "playlist error"
	errStr := "not found"

	err := NewPlaylistError(msg, errStr)

	if err == nil {
		t.Fatal("NewPlaylistError returned nil")
	}

	if !strings.Contains(err.Error(), msg) {
		t.Errorf("Error message should contain %q", msg)
	}

	if !strings.Contains(err.Error(), errStr) {
		t.Errorf("Error message should contain %q", errStr)
	}
}

func TestUserError(t *testing.T) {
	msg := "user error"
	errStr := "unauthorized"

	err := NewUserError(msg, errStr)

	if err == nil {
		t.Fatal("NewUserError returned nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestSongError(t *testing.T) {
	msg := "song error"
	errStr := "not available"

	err := NewSongError(msg, errStr)

	if err == nil {
		t.Fatal("NewSongError returned nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestArtistError(t *testing.T) {
	msg := "artist error"
	errStr := "not found"

	err := NewArtistError(msg, errStr)

	if err == nil {
		t.Fatal("NewArtistError returned nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestAlbumError(t *testing.T) {
	msg := "album error"
	errStr := "invalid id"

	err := NewAlbumError(msg, errStr)

	if err == nil {
		t.Fatal("NewAlbumError returned nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestPodcastError(t *testing.T) {
	msg := "podcast error"
	errStr := "episode unavailable"

	err := NewPodcastError(msg, errStr)

	if err == nil {
		t.Fatal("NewPodcastError returned nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestWebSocketError(t *testing.T) {
	msg := "websocket error"
	errStr := "connection closed"

	err := NewWebSocketError(msg, errStr)

	if err == nil {
		t.Fatal("NewWebSocketError returned nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestPlayerError(t *testing.T) {
	msg := "player error"
	errStr := "playback failed"

	err := NewPlayerError(msg, errStr)

	if err == nil {
		t.Fatal("NewPlayerError returned nil")
	}

	expected := msg + ": " + errStr
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestErrorWithEmptyErrString(t *testing.T) {
	// Test all error types with empty err string
	testCases := []struct {
		name string
		err  error
	}{
		{"LoginError", NewLoginError("msg", "")},
		{"RequestError", NewRequestError("msg", "")},
		{"BaseClientError", NewBaseClientError("msg", "")},
		{"PlaylistError", NewPlaylistError("msg", "")},
		{"UserError", NewUserError("msg", "")},
		{"SongError", NewSongError("msg", "")},
		{"ArtistError", NewArtistError("msg", "")},
		{"AlbumError", NewAlbumError("msg", "")},
		{"PodcastError", NewPodcastError("msg", "")},
		{"WebSocketError", NewWebSocketError("msg", "")},
		{"PlayerError", NewPlayerError("msg", "")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatal("Error constructor returned nil")
			}

			errMsg := tc.err.Error()
			if errMsg != "msg" {
				t.Errorf("Expected 'msg', got %q", errMsg)
			}
		})
	}
}

func TestAllErrorTypesImplementError(t *testing.T) {
	// Verify all error types implement the error interface
	var _ error = NewLoginError("", "")
	var _ error = NewRequestError("", "")
	var _ error = NewBaseClientError("", "")
	var _ error = NewPlaylistError("", "")
	var _ error = NewUserError("", "")
	var _ error = NewSongError("", "")
	var _ error = NewArtistError("", "")
	var _ error = NewAlbumError("", "")
	var _ error = NewPodcastError("", "")
	var _ error = NewWebSocketError("", "")
	var _ error = NewPlayerError("", "")
}