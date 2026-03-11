package spotapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spotapi/spotapi-go/internal/errors"
	"github.com/spotapi/spotapi-go/internal/http"
)

type PublicAlbum struct {
	Base      *http.BaseClient
	AlbumId   string
	AlbumLink string
}

func NewPublicAlbum(album string, client *http.Client, language string) *PublicAlbum {
	albumId := album
	if strings.Contains(album, "album/") {
		parts := strings.Split(album, "album/")
		albumId = parts[len(parts)-1]
	}

	return &PublicAlbum{
		Base:      http.NewBaseClient(client, language),
		AlbumId:   albumId,
		AlbumLink: "https://open.spotify.com/album/" + albumId,
	}
}

func (a *PublicAlbum) GetAlbumInfo(limit int, offset int) (map[string]interface{}, error) {
	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	vars, _ := json.Marshal(map[string]interface{}{
		"locale": "",
		"uri":    fmt.Sprintf("spotify:album:%s", a.AlbumId),
		"offset": offset,
		"limit":  limit,
	})

	extensions, _ := json.Marshal(map[string]interface{}{
		"persistedQuery": map[string]interface{}{
			"version":    1,
			"sha256Hash": a.Base.PartHash("getAlbum"),
		},
	})

	params := url.Values{}
	params.Set("operationName", "getAlbum")
	params.Set("variables", string(vars))
	params.Set("extensions", string(extensions))

	resp, err := a.Base.Client.Post(u+"?"+params.Encode(), true, nil, nil)
	if err != nil {
		return nil, errors.NewAlbumError("Could not get album info", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, errors.NewAlbumError("Invalid JSON", "")
}
