package spotapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spotapi/spotapi-go/internal/errors"
	"github.com/spotapi/spotapi-go/internal/http"
)

type PublicPlaylist struct {
	Base       *http.BaseClient
	PlaylistId string
}

func NewPublicPlaylist(playlist string, client *http.Client, language string) *PublicPlaylist {
	playlistId := playlist
	if strings.Contains(playlist, "playlist/") {
		parts := strings.Split(playlist, "playlist/")
		playlistId = parts[len(parts)-1]
	}

	return &PublicPlaylist{
		Base:       http.NewBaseClient(client, language),
		PlaylistId: playlistId,
	}
}

func (p *PublicPlaylist) GetPlaylistInfo(limit int, offset int) (map[string]interface{}, error) {
	apiUri := "https://api-partner.spotify.com/pathfinder/v1/query"

	vars, _ := json.Marshal(map[string]interface{}{
		"uri":                       fmt.Sprintf("spotify:playlist:%s", p.PlaylistId),
		"offset":                    offset,
		"limit":                     limit,
		"enableWatchFeedEntrypoint": false,
	})

	hash, err := p.Base.PartHash("fetchPlaylist")
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
	params.Set("operationName", "fetchPlaylist")
	params.Set("variables", string(vars))
	params.Set("extensions", string(extensions))

	resp, err := p.Base.Client.Post(apiUri+"?"+params.Encode(), true, nil, nil)
	if err != nil {
		return nil, errors.NewPlaylistError("Could not get playlist info", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, errors.NewPlaylistError("Invalid JSON", "")
}

type PrivatePlaylist struct {
	Base       *http.BaseClient
	Login      *Login
	PlaylistId string
}

func NewPrivatePlaylist(l *Login, playlist string, language string) *PrivatePlaylist {
	playlistId := playlist
	if strings.Contains(playlist, "playlist/") {
		parts := strings.Split(playlist, "playlist/")
		playlistId = parts[len(parts)-1]
	}

	return &PrivatePlaylist{
		Base:       http.NewBaseClient(l.Config.Client, language),
		Login:      l,
		PlaylistId: playlistId,
	}
}

func (p *PrivatePlaylist) CreatePlaylist(name string) (string, error) {
	apiUri := "https://spclient.wg.spotify.com/playlist/v2/playlist"
	payload := map[string]interface{}{
		"ops": []map[string]interface{}{
			{
				"kind": 6,
				"updateListAttributes": map[string]interface{}{
					"newAttributes": map[string]interface{}{
						"values": map[string]interface{}{
							"name":             name,
							"formatAttributes": []interface{}{},
							"pictureSize":      []interface{}{},
						},
						"noValue": []interface{}{},
					},
				},
			},
		},
	}

	resp, err := p.Login.Config.Client.Post(apiUri, true, nil, payload)
	if err != nil {
		return "", errors.NewPlaylistError("Could not stage create playlist", err.Error())
	}

	bodyStr, ok := resp.Body.(string)
	if !ok {
		return "", errors.NewPlaylistError("Could not stage create playlist", "Response is not a string")
	}

	if idx := strings.Index(bodyStr, "spotify:playlist:"); idx != -1 {
		return bodyStr[idx:], nil
	}

	return "", errors.NewPlaylistError("Could not find desired playlist ID", "")
}
