package spotapi

// track holds spotify track fields needed by bedrock-api.
type Track struct {
	ID          string // bare spotify id, e.g. "5nujrmhLynf4yMoMtj8AQF"
	URI         string // "spotify:track:5nujrmhLynf4yMoMtj8AQF"
	Title       string
	Artist      string   // primary artist name
	Artists     []string // all artist names
	ArtistIDs   []string // all artist ids (bare spotify ids), parallel to Artists
	AlbumID     string   // bare album id
	AlbumTitle  string
	CoverURL    string // largest available cover image
	PreviewURL  string // 30-second preview (may be empty)
	DurationMs  int64
	TrackNumber int
	Playcount   string // raw string from Spotify (can exceed int64)
	Explicit    bool
}

// artist captures a spotify artist profile.
type Artist struct {
	ID               string
	URI              string
	Name             string
	AvatarURL        string
	MonthlyListeners int64
	Followers        int64 // real follower count from stats.followers (0 in search results)
	Biography        string
}

// album represents a spotify release with its tracks.
type Album struct {
	ID          string
	URI         string
	Title       string
	Artist      string
	CoverURL    string
	ReleaseDate string // iso 8601, e.g. "2020-03-27"
	TotalTracks int
	Tracks      []Track
}

// playlist represents a spotify playlist with its tracks.
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

//  helpers for digging into raw map[string]interface{} responses

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

// artistInfo holds a spotify artist name and bare id.
type artistInfo struct {
	ID   string
	Name string
}

// artistInfos extracts ids and names from artists.items[].{uri,id,profile.name}
func artistInfos(artistsObj map[string]interface{}) []artistInfo {
	items, _ := artistsObj["items"].([]interface{})
	var out []artistInfo
	for _, it := range items {
		im, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		name := digStr(im, "profile", "name")
		if name == "" {
			continue
		}
		// prefer uri so we can strip the spotify:artist: prefix; fall back to raw id when present.
		id := digStr(im, "uri")
		if id != "" {
			id = idFromURI(id)
		}
		if id == "" {
			id = digStr(im, "id")
		}
		out = append(out, artistInfo{ID: id, Name: name})
	}
	return out
}

// artistNames is a backwards-compatible helper that returns only names from artistInfos.
func artistNames(artistsObj map[string]interface{}) []string {
	infos := artistInfos(artistsObj)
	names := make([]string, 0, len(infos))
	for _, ai := range infos {
		if ai.Name != "" {
			names = append(names, ai.Name)
		}
	}
	return names
}

// idFromURI extracts the bare spotify id from a uri like "spotify:track:xxx" or "spotify:album:xxx".
// returns the input unchanged if it contains no colon.
func idFromURI(uri string) string {
	if uri == "" {
		return ""
	}
	// find second colon
	first := -1
	for i := 0; i < len(uri); i++ {
		if uri[i] == ':' {
			if first == -1 {
				first = i
			} else {
				return uri[i+1:]
			}
		}
	}
	if first >= 0 {
		return uri[first+1:]
	}
	return uri
}

// parseTrackUnion converts any Spotify track map into a Track.
// handles both gettrack (trackunion) and embedded track objects from
// search results, album tracksv2 items, and playlist itemv2.data.
func parseTrackUnion(t map[string]interface{}) *Track {
	if t == nil {
		return nil
	}

	tr := &Track{
		ID:    digStr(t, "id"),
		URI:   digStr(t, "uri"),
		Title: digStr(t, "name"),
	}

	// album/playlist track items have no bare "id" — extract from uri
	if tr.ID == "" {
		tr.ID = idFromURI(tr.URI)
	}

	// duration: getTrack uses "duration", playlist items use "trackDuration"
	if dur := digMap(t, "duration"); dur != nil {
		tr.DurationMs = int64(digFloat(dur, "totalMilliseconds"))
	} else if dur := digMap(t, "trackDuration"); dur != nil {
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

	// artists: getTrack uses firstArtist/otherArtists;
	// search/album/playlist items use a flat "artists.items[]" object.
	if fa := digMap(t, "firstArtist"); fa != nil {
		infos := artistInfos(fa)
		if len(infos) > 0 {
			tr.Artist = infos[0].Name
			for _, ai := range infos {
				tr.Artists = append(tr.Artists, ai.Name)
				tr.ArtistIDs = append(tr.ArtistIDs, ai.ID)
			}
		}
	}
	if oa := digMap(t, "otherArtists"); oa != nil {
		infos := artistInfos(oa)
		for _, ai := range infos {
			tr.Artists = append(tr.Artists, ai.Name)
			tr.ArtistIDs = append(tr.ArtistIDs, ai.ID)
		}
	}
	// flat "artists" object used in search + album + playlist track items
	if tr.Artist == "" {
		if ar := digMap(t, "artists"); ar != nil {
			infos := artistInfos(ar)
			if len(infos) > 0 {
				tr.Artist = infos[0].Name
				for _, ai := range infos {
					tr.Artists = append(tr.Artists, ai.Name)
					tr.ArtistIDs = append(tr.ArtistIDs, ai.ID)
				}
			}
		}
	}

	// albumOfTrack
	if al := digMap(t, "albumOfTrack"); al != nil {
		tr.AlbumID = digStr(al, "id")
		if tr.AlbumID == "" {
			tr.AlbumID = idFromURI(digStr(al, "uri"))
		}
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
