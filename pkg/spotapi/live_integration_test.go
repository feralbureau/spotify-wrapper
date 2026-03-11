package spotapi

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/http"
)

func pretty(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func newLiveClient(t *testing.T) *http.Client {
	t.Helper()
	client, err := http.NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		t.Fatalf("create tls-client: %v", err)
	}
	return client
}

type liveTestResult struct {
	name     string
	passed   bool
	duration time.Duration
}

var (
	liveTestMu      sync.Mutex
	liveTestResults []liveTestResult
)

func registerLiveTest(t *testing.T) {
	start := time.Now()
	t.Cleanup(func() {
		liveTestMu.Lock()
		liveTestResults = append(liveTestResults, liveTestResult{
			name:     t.Name(),
			passed:   !t.Failed(),
			duration: time.Since(start),
		})
		liveTestMu.Unlock()
	})
}

// --- helpers to dig IDs out of raw GraphQL responses -----------------------

func firstTrackID(t *testing.T, data map[string]interface{}) string {
	t.Helper()
	defer func() { recover() }()
	items := data["data"].(map[string]interface{})["searchV2"].(map[string]interface{})["tracks"].(map[string]interface{})["items"].([]interface{})
	if len(items) == 0 {
		return ""
	}
	return items[0].(map[string]interface{})["item"].(map[string]interface{})["data"].(map[string]interface{})["id"].(string)
}

func firstArtistID(t *testing.T, data map[string]interface{}) string {
	t.Helper()
	defer func() { recover() }()
	items := data["data"].(map[string]interface{})["searchV2"].(map[string]interface{})["artists"].(map[string]interface{})["items"].([]interface{})
	if len(items) == 0 {
		return ""
	}
	return items[0].(map[string]interface{})["item"].(map[string]interface{})["data"].(map[string]interface{})["id"].(string)
}

// ---------------------------------------------------------------------------
// 0. Diagnostics — show exactly what BaseClient gets from Spotify
// ---------------------------------------------------------------------------

func TestLiveDiagSession(t *testing.T) {
	registerLiveTest(t)
	client := newLiveClient(t)

	// ---- show raw JS links from open.spotify.com ----
	rawResp, err := client.Get("https://open.spotify.com", false, nil)
	if err != nil {
		t.Fatalf("GET open.spotify.com: %v", err)
	}
	bodyStr, _ := rawResp.Body.(string)
	srcLinks := extractAttrDebug(bodyStr, `src="`)
	hrefLinks := extractAttrDebug(bodyStr, `href="`)
	t.Logf("=== src= JS links (%d) ===", len(srcLinks))
	for i, l := range srcLinks {
		t.Logf("  src[%d] %s", i, l)
	}
	t.Logf("=== href= JS links (%d) ===", len(hrefLinks))
	for i, l := range hrefLinks {
		t.Logf("  href[%d] %s", i, l)
	}

	// ---- full session init ----
	song := NewSong(nil, client, "en")
	bc := song.Base

	if err := bc.GetSession(); err != nil {
		t.Fatalf("GetSession error: %v", err)
	}
	t.Logf("ClientVersion : %s", bc.ClientVersion)
	t.Logf("ClientId      : %s", bc.ClientId)
	t.Logf("DeviceId      : %s", bc.DeviceId)
	t.Logf("AccessToken   : %.40s…", bc.AccessToken)
	t.Logf("JsPack        : %s", bc.JsPack)

	if err := bc.GetClientToken(); err != nil {
		t.Fatalf("GetClientToken error: %v", err)
	}
	t.Logf("ClientToken   : %.40s…", bc.ClientToken)

	// Check persisted query hashes
	for _, name := range []string{"searchDesktop", "getTrack", "searchArtists", "queryArtistOverview", "getAlbum", "fetchPlaylist"} {
		h, err := bc.PartHash(name)
		t.Logf("PartHash(%-25s) => %q  err=%v", name, h, err)
	}
}

func extractAttrDebug(html, attr string) []string {
	var out []string
	h := html
	for {
		idx := strings.Index(h, attr)
		if idx == -1 {
			break
		}
		h = h[idx+len(attr):]
		end := strings.IndexByte(h, '"')
		if end == -1 {
			break
		}
		link := h[:end]
		if strings.HasSuffix(link, ".js") {
			out = append(out, link)
		}
		h = h[end+1:]
	}
	return out
}

// ---------------------------------------------------------------------------
// 1. Search tracks: "Lucy Bedroque"
// ---------------------------------------------------------------------------

func TestLiveSongSearch_LucyBedroque(t *testing.T) {
	registerLiveTest(t)
	client := newLiveClient(t)
	song := NewSong(nil, client, "en")

	data, err := song.QuerySongs("Lucy Bedroque", 10, 0)
	if err != nil {
		t.Fatalf("QuerySongs error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty response")
	}
	t.Logf("=== SEARCH TRACKS 'Lucy Bedroque' ===\n%s", pretty(data))
}

// ---------------------------------------------------------------------------
// 2. Get single track info (use ID obtained from search)
// ---------------------------------------------------------------------------

func TestLiveSongGetTrackInfo_LucyBedroque(t *testing.T) {
	registerLiveTest(t)
	client := newLiveClient(t)
	song := NewSong(nil, client, "en")

	// first search to grab a real ID
	searchData, err := song.QuerySongs("Lucy Bedroque", 5, 0)
	if err != nil {
		t.Fatalf("search before GetTrackInfo: %v", err)
	}
	trackID := firstTrackID(t, searchData)
	if trackID == "" {
		t.Log("could not parse track ID from search, using fallback")
		trackID = "5nujrmhLynf4yMoMtj8AQF"
	}
	t.Logf("track ID: %s", trackID)

	data, err := song.GetTrackInfo(trackID)
	if err != nil {
		t.Fatalf("GetTrackInfo error: %v", err)
	}
	t.Logf("=== TRACK INFO id=%s ===\n%s", trackID, pretty(data))
}

// ---------------------------------------------------------------------------
// 3. Search artists: "Lucy Bedroque"
// ---------------------------------------------------------------------------

func TestLiveArtistSearch_LucyBedroque(t *testing.T) {
	registerLiveTest(t)
	client := newLiveClient(t)
	artist := NewArtist(nil, client, "en")

	data, err := artist.QueryArtists("Lucy Bedroque", 5, 0)
	if err != nil {
		t.Fatalf("QueryArtists error: %v", err)
	}
	t.Logf("=== SEARCH ARTISTS 'Lucy Bedroque' ===\n%s", pretty(data))
}

// ---------------------------------------------------------------------------
// 4. Get full artist profile (ID from search)
// ---------------------------------------------------------------------------

func TestLiveArtistGetProfile_LucyBedroque(t *testing.T) {
	registerLiveTest(t)
	client := newLiveClient(t)
	artist := NewArtist(nil, client, "en")

	searchData, err := artist.QueryArtists("Lucy Bedroque", 5, 0)
	if err != nil {
		t.Fatalf("artist search: %v", err)
	}
	artistID := firstArtistID(t, searchData)
	if artistID == "" {
		t.Log("could not parse artist ID, using fallback")
		artistID = "4MzJMcHQBl9SIYSjwWn8QW"
	}
	t.Logf("artist ID: %s", artistID)

	data, err := artist.GetArtist(artistID, "en")
	if err != nil {
		t.Fatalf("GetArtist error: %v", err)
	}
	t.Logf("=== ARTIST PROFILE id=%s ===\n%s", artistID, pretty(data))
}

// ---------------------------------------------------------------------------
// 5. Get public album info (Spotify album ID)
// ---------------------------------------------------------------------------

func TestLivePublicAlbumInfo_LucyBedroque(t *testing.T) {
	registerLiveTest(t)
	client := newLiveClient(t)

	// Lucy Bedroque / glittr – 'collected' album
	album := NewPublicAlbum("6vAMeQMOcNEPHYXJRmGp73", client, "en")
	data, err := album.GetAlbumInfo(50, 0)
	if err != nil {
		t.Fatalf("GetAlbumInfo error: %v", err)
	}
	t.Logf("=== PUBLIC ALBUM INFO ===\n%s", pretty(data))
}

// ---------------------------------------------------------------------------
// 6. Get public playlist info
// ---------------------------------------------------------------------------

func TestLivePublicPlaylistInfo(t *testing.T) {
	registerLiveTest(t)
	client := newLiveClient(t)
	playlist := NewPublicPlaylist("1j9uOH2jcv3yeNjhmPhowD", client, "en")

	data, err := playlist.GetPlaylistInfo(50, 0)
	if err != nil {
		t.Fatalf("GetPlaylistInfo error: %v", err)
	}
	t.Logf("=== PUBLIC PLAYLIST INFO ===\n%s", pretty(data))
}

func TestLiveSearchStructure(t *testing.T) {
	registerLiveTest(t)
	client := newLiveClient(t)
	song := NewSong(nil, client, "en")

	data, err := song.QuerySongs("lucy bedroque", 3, 0)
	if err != nil {
		t.Fatalf("QuerySongs: %v", err)
	}
	sv2, ok := data["data"].(map[string]interface{})["searchV2"].(map[string]interface{})
	if !ok {
		t.Fatal("no searchV2")
	}
	t.Logf("searchV2 keys: %v", func() []string {
		var ks []string
		for k := range sv2 {
			ks = append(ks, k)
		}
		return ks
	}())
	// playlists
	if plRaw, ok2 := sv2["playlists"].(map[string]interface{}); ok2 {
		items, _ := plRaw["items"].([]interface{})
		if len(items) > 0 {
			t.Logf("=== playlist item[0] ===\n%s", pretty(items[0]))
		} else {
			t.Log("no playlist items")
		}
	}
	// albumsV2
	if alRaw, ok2 := sv2["albumsV2"].(map[string]interface{}); ok2 {
		items, _ := alRaw["items"].([]interface{})
		if len(items) > 0 {
			t.Logf("=== albumsV2 item[0] ===\n%s", pretty(items[0]))
		}
	}
}

func TestLiveSummary(t *testing.T) {
	liveTestMu.Lock()
	results := append([]liveTestResult(nil), liveTestResults...)
	liveTestMu.Unlock()

	if len(results) == 0 {
		t.Skip("live tests not run")
	}

	t.Log("\n=== Live Integration Summary ===")
	allPass := true
	for _, res := range results {
		status := "PASS"
		if !res.passed {
			status = "FAIL"
			allPass = false
		}
		t.Logf("  %s  %-50s  %s", status, res.name, res.duration.Round(time.Millisecond))
	}
	if !allPass {
		t.Log("some live tests failed — check earlier logs")
	}
}
