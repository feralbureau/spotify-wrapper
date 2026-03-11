// spotapi provides a high-level go client for spotify's private partner api.
// it handles session/totp auth so oauth credentials are optional.
//
// quick start:
//
//	c, err := spotapi.NewClient("en")
//	tracks, err := c.SearchTracks("Lucy Bedroque", 10, 0)
//
// module path: github.com/spotapi/spotapi-go
package spotapi

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bogdanfinn/tls-client/profiles"
	sphttp "github.com/spotapi/spotapi-go/internal/http"
)

// client wraps shared spotapi state so the tls session and tokens stay cached.
type Client struct {
	lang   string
	http   *sphttp.Client
	song   *Song
	artist *artistService
}

// newclient builds a chrome 120 tls client; lang defaults to "en" when empty.
func NewClient(lang string) (*Client, error) {
	if lang == "" {
		lang = "en"
	}
	h, err := sphttp.NewClient(profiles.Chrome_120, "", 3)
	if err != nil {
		return nil, fmt.Errorf("spotapi: create http client: %w", err)
	}
	c := &Client{
		lang: lang,
		http: h,
	}
	c.song = NewSong(nil, h, lang)
	c.artist = NewArtist(nil, h, lang)
	return c, nil
}

//  Tracks 

// searchtracks returns up to limit tracks for a query starting at offset.
func (c *Client) SearchTracks(query string, limit, offset int) ([]Track, error) {
	raw, err := c.song.QuerySongs(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("spotapi: SearchTracks: %w", err)
	}

	tracks, _ := digMap(raw, "data", "searchV2", "tracksV2")["items"].([]interface{})
	var out []Track
	for _, item := range tracks {
		im, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		// searchV2 wraps in "item.data"
		data := digMap(im, "item", "data")
		if data == nil {
			// fallback: direct "track" key (older schema)
			data = digMap(im, "track")
		}
		if t := parseSearchTrackItem(data); t != nil && t.ID != "" {
			out = append(out, *t)
		}
	}
	return out, nil
}

// gettrack returns metadata for a single spotify id and accepts bare ids or full uris.
func (c *Client) GetTrack(id string) (*Track, error) {
	id = stripPrefix(id, "spotify:track:")
	raw, err := c.song.GetTrackInfo(id)
	if err != nil {
		return nil, fmt.Errorf("spotapi: GetTrack: %w", err)
	}
	union := digMap(raw, "data", "trackUnion")
	if union == nil {
		union = digMap(raw, "data", "track")
	}
	t := parseTrackUnion(union)
	if t == nil || t.ID == "" {
		return nil, fmt.Errorf("spotapi: GetTrack: empty response for id=%s", id)
	}
	return t, nil
}

//  Artists 

// searchartists returns up to limit artists for the query.
func (c *Client) SearchArtists(query string, limit, offset int) ([]Artist, error) {
	raw, err := c.artist.QueryArtists(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("spotapi: SearchArtists: %w", err)
	}

	items, _ := digMap(raw, "data", "searchV2", "artists")["items"].([]interface{})
	var out []Artist
	for _, item := range items {
		im, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		data := digMap(im, "item", "data")
		if data == nil {
			data = digMap(im, "data")
		}
		if a := parseArtistData(data); a != nil && a.ID != "" {
			out = append(out, *a)
		}
	}
	// enrich each artist with followers/monthly-listeners via parallel GetArtist calls,
	// since searchV2.artists items carry no stats.
	var wg sync.WaitGroup
	for i := range out {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			full, _, err := c.GetArtist(out[i].ID)
			if err == nil && full != nil {
				out[i].Followers = full.Followers
				out[i].MonthlyListeners = full.MonthlyListeners
			}
		}(i)
	}
	wg.Wait()
	return out, nil
}

// getartist fetches the artist profile and discography for a spotify id.
func (c *Client) GetArtist(id string) (*Artist, []Album, error) {
	id = stripPrefix(id, "spotify:artist:")
	raw, err := c.artist.GetArtist(id, c.lang)
	if err != nil {
		return nil, nil, fmt.Errorf("spotapi: GetArtist: %w", err)
	}
	data := digMap(raw, "data", "artistUnion")
	if data == nil {
		data = digMap(raw, "data", "artist")
	}
	a := parseArtistProfile(data)
	if a == nil || a.ID == "" {
		return nil, nil, fmt.Errorf("spotapi: GetArtist: empty response for id=%s", id)
	}
	var albums []Album
	if disc := digMap(data, "discography"); disc != nil {
		if albs := digMap(disc, "albums"); albs != nil {
			if items, ok := albs["items"].([]interface{}); ok {
				for _, it := range items {
					im, ok := it.(map[string]interface{})
					if !ok {
						continue
					}
					if releases := digMap(im, "releases"); releases != nil {
						if rItems, ok := releases["items"].([]interface{}); ok && len(rItems) > 0 {
							if r, ok := rItems[0].(map[string]interface{}); ok {
								if al := parseDiscographyAlbum(r, a.Name); al != nil && al.ID != "" {
									albums = append(albums, *al)
								}
							}
						}
					}
				}
			}
		}
	}
	return a, albums, nil
}

//  Albums 

// getalbum returns album metadata and up to limit tracks starting at offset; accepts bare album ids or spotify uris/urls.
func (c *Client) GetAlbum(id string, limit, offset int) (*Album, error) {
	id = stripPrefix(id, "spotify:album:")
	pa := NewPublicAlbum(id, c.http, c.lang)
	raw, err := pa.GetAlbumInfo(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("spotapi: GetAlbum: %w", err)
	}
	al := parseAlbumUnion(digMap(raw, "data", "albumUnion"))
	if al == nil {
		return nil, fmt.Errorf("spotapi: GetAlbum: empty response for id=%s", id)
	}
	return al, nil
}

//  Playlists 

// searchalbums returns up to limit albums for the query; totalTracks is not populated (call getalbum for full detail).
func (c *Client) SearchAlbums(query string, limit, offset int) ([]Album, error) {
	raw, err := c.song.QuerySongs(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("spotapi: SearchAlbums: %w", err)
	}
	items, _ := digMap(raw, "data", "searchV2", "albumsV2")["items"].([]interface{})
	var out []Album
	for _, item := range items {
		im, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		data := digMap(im, "data")
		if al := parseSearchAlbumItem(data); al != nil && al.ID != "" {
			out = append(out, *al)
		}
	}
	return out, nil
}

// searchplaylists returns up to limit playlists for the query; totalTracks is not populated (call getplaylist for full detail).
func (c *Client) SearchPlaylists(query string, limit, offset int) ([]Playlist, error) {
	raw, err := c.song.QuerySongs(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("spotapi: SearchPlaylists: %w", err)
	}
	items, _ := digMap(raw, "data", "searchV2", "playlists")["items"].([]interface{})
	var out []Playlist
	for _, item := range items {
		im, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		data := digMap(im, "data")
		if pl := parseSearchPlaylistItem(data); pl != nil && pl.ID != "" {
			out = append(out, *pl)
		}
	}
	return out, nil
}

// getplaylist returns playlist metadata and up to limit tracks starting at offset; accepts bare playlist ids, spotify uris, or open.spotify.com urls.
func (c *Client) GetPlaylist(id string, limit, offset int) (*Playlist, error) {
	id = stripPrefix(id, "spotify:playlist:")
	pp := NewPublicPlaylist(id, c.http, c.lang)
	raw, err := pp.GetPlaylistInfo(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("spotapi: GetPlaylist: %w", err)
	}
	pl := parsePlaylistData(digMap(raw, "data", "playlistV2"))
	if pl == nil {
		return nil, fmt.Errorf("spotapi: GetPlaylist: empty response for id=%s", id)
	}
	return pl, nil
}

//  internal helpers 

func stripPrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// parseSearchAlbumItem parses a searchV2.albumsV2.items[].data map into an Album.
func parseSearchAlbumItem(d map[string]interface{}) *Album {
	if d == nil {
		return nil
	}
	// skip non-album items (e.g. pre-releases, compilations with different __typename)
	if tn, _ := d["__typename"].(string); tn != "" && tn != "Album" {
		return nil
	}
	al := &Album{
		URI:   digStr(d, "uri"),
		Title: digStr(d, "name"),
	}
	al.ID = digStr(d, "id")
	if al.ID == "" {
		al.ID = idFromURI(al.URI)
	}
	if artists := digMap(d, "artists"); artists != nil {
		names := artistNames(artists)
		if len(names) > 0 {
			al.Artist = names[0]
		}
	}
	if ca := digMap(d, "coverArt"); ca != nil {
		al.CoverURL = bestCover(ca)
	}
	if date := digMap(d, "date"); date != nil {
		al.ReleaseDate = digStr(date, "isoString")
		if al.ReleaseDate == "" {
			if y, ok := date["year"].(float64); ok {
				al.ReleaseDate = fmt.Sprintf("%d", int(y))
			}
		}
	}
	return al
}

// parseSearchPlaylistItem parses a searchV2.playlists.items[].data map into a Playlist.
func parseSearchPlaylistItem(d map[string]interface{}) *Playlist {
	if d == nil {
		return nil
	}
	pl := &Playlist{
		URI:         digStr(d, "uri"),
		Title:       digStr(d, "name"),
		Description: digStr(d, "description"),
		Owner:       digStr(d, "ownerV2", "data", "name"),
	}
	pl.ID = idFromURI(pl.URI)
	if images := digMap(d, "images"); images != nil {
		if items, ok := images["items"].([]interface{}); ok && len(items) > 0 {
			if first, ok := items[0].(map[string]interface{}); ok {
				if srcs, ok := first["sources"].([]interface{}); ok && len(srcs) > 0 {
					if s, ok := srcs[0].(map[string]interface{}); ok {
						pl.CoverURL, _ = s["url"].(string)
					}
				}
			}
		}
	}
	return pl
}

// parseDiscographyAlbum parses an artistUnion.discography.albums.items[].releases.items[0] map.
func parseDiscographyAlbum(d map[string]interface{}, artistName string) *Album {
	if d == nil {
		return nil
	}
	al := &Album{
		ID:     digStr(d, "id"),
		URI:    digStr(d, "uri"),
		Title:  digStr(d, "name"),
		Artist: artistName,
	}
	if al.ID == "" {
		al.ID = idFromURI(al.URI)
	}
	if ca := digMap(d, "coverArt"); ca != nil {
		al.CoverURL = bestCover(ca)
	}
	if date := digMap(d, "date"); date != nil {
		al.ReleaseDate = digStr(date, "isoString")
		if al.ReleaseDate == "" {
			if y, ok := date["year"].(float64); ok {
				al.ReleaseDate = fmt.Sprintf("%d", int(y))
			}
		}
	}
	if tracks := digMap(d, "tracks"); tracks != nil {
		if tc, ok := tracks["totalCount"].(float64); ok {
			al.TotalTracks = int(tc)
		}
	}
	return al
}

// parseArtistData parses the searchV2 artist data item.
func parseArtistData(d map[string]interface{}) *Artist {
	if d == nil {
		return nil
	}
	a := &Artist{
		ID:   digStr(d, "id"),
		URI:  digStr(d, "uri"),
		Name: digStr(d, "profile", "name"),
	}
	// artist search items have no bare "id" — extract from uri
	if a.ID == "" {
		a.ID = idFromURI(a.URI)
	}
	if vis := digMap(d, "visuals"); vis != nil {
		if av := digMap(vis, "avatarImage"); av != nil {
			a.AvatarURL = bestCover(av)
		}
	}
	return a
}

// parseArtistProfile parses the queryArtistOverview artistUnion.
func parseArtistProfile(d map[string]interface{}) *Artist {
	if d == nil {
		return nil
	}
	a := &Artist{
		ID:   digStr(d, "id"),
		URI:  digStr(d, "uri"),
		Name: digStr(d, "profile", "name"),
	}
	if a.ID == "" {
		a.ID = idFromURI(a.URI)
	}
	if bio := digMap(d, "profile"); bio != nil {
		a.Biography = digStr(bio, "biography", "text")
	}
	if stats := digMap(d, "stats"); stats != nil {
		if ml, ok := stats["monthlyListeners"].(float64); ok {
			a.MonthlyListeners = int64(ml)
		}
		if fl, ok := stats["followers"].(float64); ok {
			a.Followers = int64(fl)
		}
	}
	if vis := digMap(d, "visuals"); vis != nil {
		if av := digMap(vis, "avatarImage"); av != nil {
			a.AvatarURL = bestCover(av)
		}
	}
	return a
}

// parseAlbumUnion converts albumUnion map into an Album.
func parseAlbumUnion(d map[string]interface{}) *Album {
	if d == nil {
		return nil
	}
	if tn, _ := d["__typename"].(string); tn == "NotFound" {
		return nil
	}
	al := &Album{
		ID:    digStr(d, "id"),
		URI:   digStr(d, "uri"),
		Title: digStr(d, "name"),
	}
	if al.ID == "" {
		al.ID = idFromURI(al.URI)
	}
	if artists := digMap(d, "artists"); artists != nil {
		names := artistNames(artists)
		if len(names) > 0 {
			al.Artist = names[0]
		}
	}
	if ca := digMap(d, "coverArt"); ca != nil {
		al.CoverURL = bestCover(ca)
	}
	if date := digMap(d, "date"); date != nil {
		al.ReleaseDate = digStr(date, "isoString")
		if al.ReleaseDate == "" {
			if y, ok := date["year"].(float64); ok {
				al.ReleaseDate = fmt.Sprintf("%d", int(y))
			}
		}
	}
	if tracks := digMap(d, "tracksV2"); tracks != nil {
		if tc, ok := tracks["totalCount"].(float64); ok {
			al.TotalTracks = int(tc)
		}
		if items, ok := tracks["items"].([]interface{}); ok {
			for _, it := range items {
				im, ok := it.(map[string]interface{})
				if !ok {
					continue
				}
				trackObj := digMap(im, "track")
				if trackObj == nil {
					continue
				}
				t := parseTrackUnion(trackObj)
				if t == nil {
					continue
				}
				// backfill album-level fields that may be absent inside album track items
				if t.AlbumID == "" {
					t.AlbumID = al.ID
				}
				if t.AlbumTitle == "" {
					t.AlbumTitle = al.Title
				}
				if t.CoverURL == "" {
					t.CoverURL = al.CoverURL
				}
				al.Tracks = append(al.Tracks, *t)
			}
		}
	}
	return al
}

// parsePlaylistData converts a playlistV2 map into a Playlist.
func parsePlaylistData(d map[string]interface{}) *Playlist {
	if d == nil {
		return nil
	}
	pl := &Playlist{
		ID:          digStr(d, "id"),
		URI:         digStr(d, "uri"),
		Title:       digStr(d, "name"),
		Description: digStr(d, "description"),
		Owner:       digStr(d, "ownerV2", "data", "name"),
	}
	if images := digMap(d, "images"); images != nil {
		if items, ok := images["items"].([]interface{}); ok && len(items) > 0 {
			if first, ok := items[0].(map[string]interface{}); ok {
				if srcs, ok := first["sources"].([]interface{}); ok && len(srcs) > 0 {
					if s, ok := srcs[0].(map[string]interface{}); ok {
						pl.CoverURL, _ = s["url"].(string)
					}
				}
			}
		}
	}
	if content := digMap(d, "content"); content != nil {
		if tc, ok := content["totalCount"].(float64); ok {
			pl.TotalTracks = int(tc)
		}
		if items, ok := content["items"].([]interface{}); ok {
			for _, it := range items {
				im, ok := it.(map[string]interface{})
				if !ok {
					continue
				}
				trackData := digMap(im, "itemV2", "data")
				if trackData == nil {
					continue
				}
				t := parseTrackUnion(trackData)
				if t != nil && t.ID != "" {
					pl.Tracks = append(pl.Tracks, *t)
				}
			}
		}
	}
	return pl
}
