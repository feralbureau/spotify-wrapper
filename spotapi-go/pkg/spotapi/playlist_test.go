package spotapi

import (
	"testing"
	"github.com/spotapi/spotapi-go/internal/http"
	"github.com/bogdanfinn/tls-client/profiles"
)

func TestNewPublicPlaylist(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	playlist := NewPublicPlaylist("37i9dQZF1DXcBWIGoYBM5M", client, "en")

	if playlist.PlaylistId != "37i9dQZF1DXcBWIGoYBM5M" {
		t.Errorf("Expected playlist ID 37i9dQZF1DXcBWIGoYBM5M, got %s", playlist.PlaylistId)
	}
}
