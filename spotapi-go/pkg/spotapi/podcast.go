package spotapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spotapi/spotapi-go/internal/errors"
	"github.com/spotapi/spotapi-go/internal/http"
)

type Podcast struct {
	Base      *http.BaseClient
	PodcastId string
	PodcastLink string
}

// NewPodcast creates a Podcast configured for the given podcast identifier or Spotify show URL.
// If the input contains "show/" the substring after the last "show/" is used as the podcast ID.
// The returned Podcast has Base initialized with the provided HTTP client and language, PodcastId
// set to the extracted id, and PodcastLink set to "https://open.spotify.com/show/<podcastId>".
func NewPodcast(podcast string, client *http.Client, language string) *Podcast {
	podcastId := podcast
	if strings.Contains(podcast, "show/") {
		parts := strings.Split(podcast, "show/")
		podcastId = parts[len(parts)-1]
	}

	return &Podcast{
		Base:      http.NewBaseClient(client, language),
		PodcastId: podcastId,
		PodcastLink: "https://open.spotify.com/show/" + podcastId,
	}
}

func (p *Podcast) GetEpisode(episodeId string) (map[string]interface{}, error) {
	u := "https://api-partner.spotify.com/pathfinder/v1/query"

	vars, _ := json.Marshal(map[string]interface{}{
		"uri": fmt.Sprintf("spotify:episode:%s", episodeId),
	})

	extensions, _ := json.Marshal(map[string]interface{}{
		"persistedQuery": map[string]interface{}{
			"version":    1,
			"sha256Hash": p.Base.PartHash("getEpisodeOrChapter"),
		},
	})

	params := url.Values{}
	params.Set("operationName", "getEpisodeOrChapter")
	params.Set("variables", string(vars))
	params.Set("extensions", string(extensions))

	resp, err := p.Base.Client.Post(u+"?"+params.Encode(), true, nil, nil)
	if err != nil {
		return nil, errors.NewPodcastError("Could not get episode info", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, errors.NewPodcastError("Invalid JSON", "")
}
