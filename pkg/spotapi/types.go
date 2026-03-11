package spotapi

// Track is a single Spotify track with all fields needed by bedrock-api.
type Track struct {
	ID          string   // bare Spotify ID, e.g. "5nujrmhLynf4yMoMtj8AQF"
	URI         string   // "spotify:track:5nujrmhLynf4yMoMtj8AQF"
	Title       string
	Artist      string   // primary artist name
	Artists     []string // all artist names
	AlbumID     string   // bare album ID
	AlbumTitle  string
	CoverURL    string // largest available cover image
	PreviewURL  string // 30-second preview (may be empty)
	DurationMs  int64
	TrackNumber int
	Playcount   string // raw string from Spotify (can exceed int64)
	Explicit    bool
}

// Artist is a Spotify artist profile.
type Artist struct {
	ID               string
	URI              string
	Name             string
	AvatarURL        string
	MonthlyListeners int64
	Biography        string
}

// Album is a Spotify album with its track listing.
type Album struct {
	ID          string
	URI         string
	Title       string
	Artist      string
	CoverURL    string
	ReleaseDate string // ISO 8601, e.g. "2020-03-27"
	TotalTracks int
	Tracks      []Track
}

// Playlist is a Spotify playlist with its track listing.
type Playlist struct {
	ID          string
	URI         string
	Title       string
	Owner       string
	Description string
	CoverURL    string
	TotalTracks int
	Tracks      []Track
}

// ───── helpers for digging into raw map[string]interface{} responses ─────

func digMap(m map[string]interface{}, keys ...string) map[string]interface{} {
	cur := m
	for _, k := range keys {
		next, ok := cur[k].(map[string]interface{})
		if !ok {
			return nil
		}
		cur = next
	}
	return cur
}

func digStr(m map[string]interface{}, keys ...string) string {
	if len(keys) == 0 {
		return ""
	}
	for i := 0; i < len(keys)-1; i++ {
		next, ok := m[keys[i]].(map[string]interface{})
		if !ok {
			return ""
		}
		m = next
	}
	s, _ := m[keys[len(keys)-1]].(string)
	return s
}

func digFloat(m map[string]interface{}, keys ...string) float64 {
	if len(keys) == 0 {
		return 0
	}
	for i := 0; i < len(keys)-1; i++ {
		next, ok := m[keys[i]].(map[string]interface{})
		if !ok {
			return 0
		}
		m = next
	}
	f, _ := m[keys[len(keys)-1]].(float64)
	return f
}

// bestCover returns the URL of the widest cover art source.
func bestCover(coverArt map[string]interface{}) string {
	sources, ok := digMap(coverArt, "sources")["sources"].([]interface{})
	if !ok {
		// coverArt *is* the object with "sources" key
		raw, ok2 := coverArt["sources"].([]interface{})
		if !ok2 {
			return ""
		}
		sources = raw
	}
	best := ""
	bestW := float64(0)
	for _, s := range sources {
		sm, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		w, _ := sm["width"].(float64)
		if w > bestW {
			bestW = w
			best, _ = sm["url"].(string)
		}
	}
	return best
}

// artistNames extracts names from artists.items[].profile.name
func artistNames(artistsObj map[string]interface{}) []string {
	items, _ := artistsObj["items"].([]interface{})
	var names []string
	for _, it := range items {
		im, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		if n := digStr(im, "profile", "name"); n != "" {
			names = append(names, n)
		}
	}
	return names
}

// parseTrackUnion converts a trackUnion map (from getTrack) into a Track.
func parseTrackUnion(t map[string]interface{}) *Track {
	if t == nil {
		return nil
	}

	tr := &Track{
		ID:   digStr(t, "id"),
		URI:  digStr(t, "uri"),
		Title: digStr(t, "name"),
	}

	// duration
	if dur := digMap(t, "duration"); dur != nil {
		tr.DurationMs = int64(digFloat(dur, "totalMilliseconds"))
	}

	// track number
	if tn, ok := t["trackNumber"].(float64); ok {
		tr.TrackNumber = int(tn)
	}

	// playcount
	tr.Playcount, _ = t["playcount"].(string)

	// explicit
	if cr := digMap(t, "contentRating"); cr != nil {
		tr.Explicit = digStr(cr, "label") == "EXPLICIT"
	}

	// preview
	if prev := digMap(t, "previews"); prev != nil {
		if ap := digMap(prev, "audioPreviews"); ap != nil {
			if items, ok := ap["items"].([]interface{}); ok && len(items) > 0 {
				if im, ok := items[0].(map[string]interface{}); ok {
					tr.PreviewURL, _ = im["url"].(string)
				}
			}
		}
	}

	// firstArtist + otherArtists
	if fa := digMap(t, "firstArtist"); fa != nil {
		names := artistNames(fa)
		if len(names) > 0 {
			tr.Artist = names[0]
			tr.Artists = append(tr.Artists, names...)
		}
	}
	if oa := digMap(t, "otherArtists"); oa != nil {
		tr.Artists = append(tr.Artists, artistNames(oa)...)
	}

	// albumOfTrack
	if al := digMap(t, "albumOfTrack"); al != nil {
		tr.AlbumID = digStr(al, "id")
		tr.AlbumTitle = digStr(al, "name")
		if ca := digMap(al, "coverArt"); ca != nil {
			tr.CoverURL = bestCover(ca)
		}
	}

	return tr
}

// parseSearchTrackItem converts a searchV2.tracks.items[].item.data map into a Track.
func parseSearchTrackItem(data map[string]interface{}) *Track {
	return parseTrackUnion(data)
}
