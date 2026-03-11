package spotapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spotapi/spotapi-go/internal/errors"
	"github.com/spotapi/spotapi-go/internal/http"
)

type artistService struct {
	Base  *http.BaseClient
	login bool
}

func NewArtist(l *Login, client *http.Client, language string) *artistService {
	var bc *http.BaseClient
	isLoggedIn := false
	if l != nil {
		bc = http.NewBaseClient(l.Config.Client, language)
		isLoggedIn = l.Authorized
	} else {
		bc = http.NewBaseClient(client, language)
	}

	return &artistService{
		Base:  bc,
		login: isLoggedIn,
	}
}

func (a *artistService) QueryArtists(query string, limit int, offset int) (map[string]interface{}, error) {
	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	vars, _ := json.Marshal(map[string]interface{}{
		"searchTerm":         query,
		"offset":             offset,
		"limit":              limit,
		"numberOfTopResults": 5,
		"includeAudiobooks":  true,
		"includePreReleases": false,
	})

	hash, err := a.Base.PartHash("searchArtists")
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
	params.Set("operationName", "searchArtists")
	params.Set("variables", string(vars))
	params.Set("extensions", string(extensions))

	resp, err := a.Base.Client.Post(u+"?"+params.Encode(), true, nil, nil)
	if err != nil {
		return nil, errors.NewArtistError("Could not get artists", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, errors.NewArtistError("Invalid JSON", fmt.Sprintf("body=%v", resp.Body))
}

func (a *artistService) GetArtist(artistId string, localeCode string) (map[string]interface{}, error) {
	if strings.Contains(artistId, "artist:") {
		artistId = strings.Split(artistId, "artist:")[1]
	}

	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	vars, _ := json.Marshal(map[string]interface{}{
		"uri":    fmt.Sprintf("spotify:artist:%s", artistId),
		"locale": localeCode,
	})

	hash, err := a.Base.PartHash("queryArtistOverview")
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
	params.Set("operationName", "queryArtistOverview")
	params.Set("variables", string(vars))
	params.Set("extensions", string(extensions))

	resp, err := a.Base.Client.Get(u+"?"+params.Encode(), true, nil)
	if err != nil {
		return nil, errors.NewArtistError("Could not get artist by ID", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, errors.NewArtistError("Invalid JSON response", fmt.Sprintf("body=%v", resp.Body))
}

func (a *artistService) Follow(artistId string) error {
	return a.doFollow(artistId, "addToLibrary")
}

func (a *artistService) Unfollow(artistId string) error {
	return a.doFollow(artistId, "removeFromLibrary")
}

func (a *artistService) doFollow(artistId string, action string) error {
	if !a.login {
		return fmt.Errorf("must be logged in")
	}

	if strings.Contains(artistId, "artist:") {
		artistId = strings.Split(artistId, "artist:")[1]
	}

	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	hash, err := a.Base.PartHash(action)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"variables": map[string]interface{}{
			"uris": []string{fmt.Sprintf("spotify:artist:%s", artistId)},
		},
		"operationName": action,
		"extensions": map[string]interface{}{
			"persistedQuery": map[string]interface{}{
				"version":    1,
				"sha256Hash": hash,
			},
		},
	}

	resp, err := a.Base.Client.Post(u, true, nil, payload)
	if err != nil {
		return errors.NewArtistError("Could not follow artist", err.Error())
	}

	if resp.Fail {
		return errors.NewArtistError("Could not follow artist", "Request failed")
	}

	return nil
}
