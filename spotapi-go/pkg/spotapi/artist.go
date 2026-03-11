package spotapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spotapi/spotapi-go/internal/errors"
	"github.com/spotapi/spotapi-go/internal/http"
)

type Artist struct {
	Base  *http.BaseClient
	login bool
}

// NewArtist creates and returns an Artist client configured for the given language.
// If a non-nil Login is provided, NewArtist uses the Login's configured HTTP client and sets the client's authenticated state to the Login's Authorized value; otherwise it uses the supplied http.Client.
func NewArtist(l *Login, client *http.Client, language string) *Artist {
	var bc *http.BaseClient
	isLoggedIn := false
	if l != nil {
		bc = http.NewBaseClient(l.Config.Client, language)
		isLoggedIn = l.Authorized
	} else {
		bc = http.NewBaseClient(client, language)
	}

	return &Artist{
		Base:  bc,
		login: isLoggedIn,
	}
}

func (a *Artist) QueryArtists(query string, limit int, offset int) (map[string]interface{}, error) {
	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	vars, _ := json.Marshal(map[string]interface{}{
		"searchTerm":         query,
		"offset":             offset,
		"limit":              limit,
		"numberOfTopResults": 5,
		"includeAudiobooks":  true,
		"includePreReleases": false,
	})

	extensions, _ := json.Marshal(map[string]interface{}{
		"persistedQuery": map[string]interface{}{
			"version":    1,
			"sha256Hash": a.Base.PartHash("searchArtists"),
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

	return nil, errors.NewArtistError("Invalid JSON", "")
}

func (a *Artist) GetArtist(artistId string, localeCode string) (map[string]interface{}, error) {
	if strings.Contains(artistId, "artist:") {
		artistId = strings.Split(artistId, "artist:")[1]
	}

	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	vars, _ := json.Marshal(map[string]interface{}{
		"uri":    fmt.Sprintf("spotify:artist:%s", artistId),
		"locale": localeCode,
	})

	extensions, _ := json.Marshal(map[string]interface{}{
		"persistedQuery": map[string]interface{}{
			"version":    1,
			"sha256Hash": a.Base.PartHash("queryArtistOverview"),
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

	return nil, errors.NewArtistError("Invalid JSON response", "")
}

func (a *Artist) Follow(artistId string) error {
	return a.doFollow(artistId, "addToLibrary")
}

func (a *Artist) Unfollow(artistId string) error {
	return a.doFollow(artistId, "removeFromLibrary")
}

func (a *Artist) doFollow(artistId string, action string) error {
	if !a.login {
		return fmt.Errorf("must be logged in")
	}

	if strings.Contains(artistId, "artist:") {
		artistId = strings.Split(artistId, "artist:")[1]
	}

	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	payload := map[string]interface{}{
		"variables": map[string]interface{}{
			"uris": []string{fmt.Sprintf("spotify:artist:%s", artistId)},
		},
		"operationName": action,
		"extensions": map[string]interface{}{
			"persistedQuery": map[string]interface{}{
				"version":    1,
				"sha256Hash": a.Base.PartHash(action),
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
