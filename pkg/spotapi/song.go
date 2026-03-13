package spotapi

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spotapi/spotapi-go/internal/errors"
	"github.com/spotapi/spotapi-go/internal/http"
)

type Song struct {
	Playlist *PrivatePlaylist
	Base     *http.BaseClient
}

func NewSong(playlist *PrivatePlaylist, client *http.Client, language string) *Song {
	var bc *http.BaseClient
	if playlist != nil {
		bc = http.NewBaseClient(playlist.Login.Config.Client, language)
	} else {
		bc = http.NewBaseClient(client, language)
	}

	return &Song{
		Playlist: playlist,
		Base:     bc,
	}
}

func (s *Song) GetTrackInfo(trackId string) (map[string]interface{}, error) {
	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	vars, _ := json.Marshal(map[string]interface{}{
		"uri": fmt.Sprintf("spotify:track:%s", trackId),
	})

	hash, err := s.Base.PartHash("getTrack")
	if err != nil {
		return nil, err
	}

	extensions, _ := json.Marshal(map[string]interface{}{
		"persistedQuery": map[string]interface{}{
			"version":    1,
			"sha256Hash": hash,
		},
	})

	params := url.Values{}
	params.Set("operationName", "getTrack")
	params.Set("variables", string(vars))
	params.Set("extensions", string(extensions))

	resp, err := s.Base.Client.Post(u+"?"+params.Encode(), true, nil, nil)
	if err != nil {
		return nil, errors.NewSongError("Could not get song info", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, errors.NewSongError("Invalid JSON", fmt.Sprintf("body=%v", resp.Body))
}

func (s *Song) QuerySongs(query string, limit int, offset int) (map[string]interface{}, error) {
	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	vars, _ := json.Marshal(map[string]interface{}{
		"searchTerm":                    query,
		"offset":                        offset,
		"limit":                         limit,
		"numberOfTopResults":            5,
		"includeAudiobooks":             true,
		"includeArtistHasConcertsField": false,
		"includePreReleases":            true,
		"includeLocalConcertsField":     false,
	})

	hash, err := s.Base.PartHash("searchDesktop")
	if err != nil {
		return nil, err
	}

	extensions, _ := json.Marshal(map[string]interface{}{
		"persistedQuery": map[string]interface{}{
			"version":    1,
			"sha256Hash": hash,
		},
	})

	params := url.Values{}
	params.Set("operationName", "searchDesktop")
	params.Set("variables", string(vars))
	params.Set("extensions", string(extensions))

	resp, err := s.Base.Client.Post(u+"?"+params.Encode(), true, nil, nil)
	if err != nil {
		return nil, errors.NewSongError("Could not get songs", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, errors.NewSongError("Invalid JSON", fmt.Sprintf("body=%v", resp.Body))
}

func (s *Song) AddSongsToPlaylist(songIds []string) error {
	if s.Playlist == nil || s.Playlist.PlaylistId == "" {
		return fmt.Errorf("playlist not set")
	}

	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	uris := make([]string, len(songIds))
	for i, id := range songIds {
		uris[i] = fmt.Sprintf("spotify:track:%s", id)
	}

	hash, err := s.Base.PartHash("addToPlaylist")
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"variables": map[string]interface{}{
			"uris":        uris,
			"playlistUri": fmt.Sprintf("spotify:playlist:%s", s.Playlist.PlaylistId),
			"newPosition": map[string]interface{}{"moveType": "BOTTOM_OF_PLAYLIST", "fromUid": nil},
		},
		"operationName": "addToPlaylist",
		"extensions": map[string]interface{}{
			"persistedQuery": map[string]interface{}{
				"version":    1,
				"sha256Hash": hash,
			},
		},
	}

	_, err = s.Base.Client.Post(u, true, nil, payload)
	if err != nil {
		return errors.NewSongError("Could not add songs to playlist", err.Error())
	}

	return nil
}
