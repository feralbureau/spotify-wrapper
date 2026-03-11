package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spotapi/spotapi-go/internal/errors"
	"github.com/spotapi/spotapi-go/internal/totp"
	"github.com/spotapi/spotapi-go/internal/utils"
)

type BaseClient struct {
	Client        *Client
	JsPack        string
	AllJsPacks    []string
	ClientVersion string
	AccessToken   string
	ClientToken   string
	ClientId      string
	DeviceId      string
	RawHashes     string
	Language      string
	ServerCfg     map[string]interface{}
}

func NewBaseClient(client *Client, language string) *BaseClient {
	bc := &BaseClient{
		Client:   client,
		Language: language,
	}

	client.Authenticate = bc.AuthRule
	return bc
}

func (bc *BaseClient) AuthRule(headers map[string]string) (map[string]string, error) {
	if bc.ClientToken == "" {
		if err := bc.GetClientToken(); err != nil {
			return nil, err
		}
	}

	if bc.AccessToken == "" {
		if err := bc.GetSession(); err != nil {
			return nil, err
		}
	}

	if headers == nil {
		headers = make(map[string]string)
	}

	headers["Authorization"] = "Bearer " + bc.AccessToken
	headers["Client-Token"] = bc.ClientToken
	headers["Spotify-App-Version"] = bc.ClientVersion
	headers["Accept"] = "application/json"
	headers["Accept-Language"] = bc.Language
	headers["Content-Type"] = "application/json;charset=UTF-8"
	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	headers["app-platform"] = "WebPlayer"
	headers["Origin"] = "https://open.spotify.com"
	headers["Referer"] = "https://open.spotify.com/"

	return headers, nil
}

func (bc *BaseClient) GetSession() error {
	desktopHeaders := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Sec-Fetch-Dest":  "document",
		"Sec-Fetch-Mode":  "navigate",
		"Sec-Fetch-Site":  "none",
		"Sec-Fetch-User":  "?1",
	}
	resp, err := bc.Client.Get("https://open.spotify.com", false, desktopHeaders)
	if err != nil {
		return errors.NewBaseClientError("Could not get session", err.Error())
	}

	bodyStr, ok := resp.Body.(string)
	if !ok {
		return errors.NewBaseClientError("Could not get session", "Response body is not a string")
	}

	links := utils.ExtractJSLinks(bodyStr)
	// isJunk skips tracking scripts that never carry spotify graphql hashes.
	isJunk := func(link string) bool {
		return strings.Contains(link, "vendor~") ||
			strings.Contains(link, "gtm.") ||
			strings.Contains(link, "retargeting")
	}
	// pass 1: exact historic desktop pattern
	for _, link := range links {
		if strings.Contains(link, "web-player/web-player") && !isJunk(link) {
			bc.JsPack = link
			break
		}
	}
	// pass 2: any spotifycdn.com web-player bundle
	if bc.JsPack == "" {
		for _, link := range links {
			if strings.Contains(link, "spotifycdn.com") && strings.Contains(link, "web-player") && !isJunk(link) {
				bc.JsPack = link
				break
			}
		}
	}
	// pass 3: any spotifycdn.com js that is not junk
	if bc.JsPack == "" {
		for _, link := range links {
			if strings.Contains(link, "spotifycdn.com") && !isJunk(link) {
				bc.JsPack = link
				break
			}
		}
	}

	// store all non-junk bundles for hash search
	bc.AllJsPacks = bc.AllJsPacks[:0]
	for _, link := range links {
		if !isJunk(link) && link != bc.JsPack {
			bc.AllJsPacks = append(bc.AllJsPacks, link)
		}
	}

	// extract appServerConfig
	parts := strings.Split(bodyStr, "<script id=\"appServerConfig\" type=\"text/plain\">")
	if len(parts) > 1 {
		configPart := strings.Split(parts[1], "</script>")[0]
		decoded, err := base64.StdEncoding.DecodeString(configPart)
		if err != nil {
			return errors.NewBaseClientError("Could not decode appServerConfig", err.Error())
		}
		if err := json.Unmarshal(decoded, &bc.ServerCfg); err != nil {
			return errors.NewBaseClientError("Could not unmarshal appServerConfig", err.Error())
		}

		v, ok := bc.ServerCfg["clientVersion"].(string)
		if !ok || v == "" {
			return errors.NewBaseClientError("Could not get session", "clientVersion is missing or invalid in appServerConfig")
		}
		bc.ClientVersion = v
	} else {
		return errors.NewBaseClientError("Could not get session", "appServerConfig script tag not found")
	}

	// device id from cookie
	foundSpT := false
	for _, cookie := range resp.Raw.Cookies() {
		if cookie.Name == "sp_t" {
			bc.DeviceId = cookie.Value
			foundSpT = true
			break
		}
	}

	if !foundSpT || bc.DeviceId == "" {
		return errors.NewBaseClientError("Could not get session", "sp_t cookie (DeviceId) is missing")
	}

	return bc.GetAuthVars()
}

func (bc *BaseClient) GetAuthVars() error {
	if bc.AccessToken == "" || bc.ClientId == "" {
		t, v := totp.GenerateTOTP()
		url := fmt.Sprintf("https://open.spotify.com/api/token?reason=init&productType=web-player&totp=%s&totpVer=%d&totpServer=%s", t, v, t)

		resp, err := bc.Client.Get(url, false, nil)
		if err != nil {
			return errors.NewBaseClientError("Could not get session auth tokens", err.Error())
		}

		if data, ok := resp.Body.(map[string]interface{}); ok {
			at, ok1 := data["accessToken"].(string)
			ci, ok2 := data["clientId"].(string)
			if !ok1 || !ok2 || at == "" || ci == "" {
				return errors.NewBaseClientError("Could not get session auth tokens", "Invalid response format or empty tokens")
			}
			bc.AccessToken = at
			bc.ClientId = ci
		} else {
			return errors.NewBaseClientError("Could not get session auth tokens", "Response is not a map")
		}
	}
	return nil
}

func (bc *BaseClient) GetClientToken() error {
	if bc.ClientId == "" || bc.DeviceId == "" || bc.ClientVersion == "" {
		if err := bc.GetSession(); err != nil {
			return err
		}
	}

	url := "https://clienttoken.spotify.com/v1/clienttoken"
	payload := map[string]interface{}{
		"client_data": map[string]interface{}{
			"client_version": bc.ClientVersion,
			"client_id":      bc.ClientId,
			"js_sdk_data": map[string]interface{}{
				"device_brand": "unknown",
				"device_model": "unknown",
				"os":           "windows",
				"os_version":   "NT 10.0",
				"device_id":    bc.DeviceId,
				"device_type":  "computer",
			},
		},
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	resp, err := bc.Client.Post(url, false, headers, payload)
	if err != nil {
		return errors.NewBaseClientError("Could not get client token", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		if gt, ok := data["granted_token"].(map[string]interface{}); ok {
			token, ok := gt["token"].(string)
			if !ok || token == "" {
				return errors.NewBaseClientError("Could not get client token", "Token is missing or empty")
			}
			bc.ClientToken = token
		} else {
			return errors.NewBaseClientError("Could not get client token", "granted_token is missing or invalid")
		}
	} else {
		return errors.NewBaseClientError("Could not get client token", "Response is not a map")
	}

	return nil
}

func (bc *BaseClient) PartHash(name string) (string, error) {
	if bc.RawHashes == "" {
		if err := bc.GetSha256Hash(); err != nil {
			return "", err
		}
	}

	// simplified hash extraction
	searchQuery := fmt.Sprintf("\"%s\",\"query\",\"", name)
	if idx := strings.Index(bc.RawHashes, searchQuery); idx != -1 {
		start := idx + len(searchQuery)
		end := strings.Index(bc.RawHashes[start:], "\"")
		if end == -1 || start+end > len(bc.RawHashes) {
			return "", nil
		}
		return bc.RawHashes[start : start+end], nil
	}

	searchMutation := fmt.Sprintf("\"%s\",\"mutation\",\"", name)
	if idx := strings.Index(bc.RawHashes, searchMutation); idx != -1 {
		start := idx + len(searchMutation)
		end := strings.Index(bc.RawHashes[start:], "\"")
		if end == -1 || start+end > len(bc.RawHashes) {
			return "", nil
		}
		return bc.RawHashes[start : start+end], nil
	}

	return "", nil
}

func (bc *BaseClient) GetSha256Hash() error {
	if bc.JsPack == "" {
		if err := bc.GetSession(); err != nil {
			return err
		}
	}

	if bc.JsPack == "" {
		return errors.NewBaseClientError("Could not find JS pack", "")
	}

	resp, err := bc.Client.Get(bc.JsPack, false, nil)
	if err != nil {
		return errors.NewBaseClientError("Could not get general hashes", err.Error())
	}

	bodyStr, ok := resp.Body.(string)
	if !ok {
		return errors.NewBaseClientError("Could not get general hashes", "JS pack response body is not a string")
	}
	bc.RawHashes = bodyStr

	// also load any other known bundles (e.g. vendor bundle) to widen hash search
	for _, extra := range bc.AllJsPacks {
		if eResp, eErr := bc.Client.Get(extra, false, nil); eErr == nil {
			if bStr, ok := eResp.Body.(string); ok {
				bc.RawHashes += bStr
			}
		}
	}

	// derive base url from the js pack path (e.g. mobile-web-player/ or web-player/)
	baseURL := "https://open.spotifycdn.com/cdn/build/web-player/"
	if idx := strings.LastIndex(bc.JsPack, "/"); idx != -1 {
		baseURL = bc.JsPack[:idx+1]
	}

	m1, m2 := utils.ExtractMappings(bc.RawHashes)
	if m1 == nil || m2 == nil {
		return nil // some packs don't carry chunk maps inline
	}

	urls := utils.CombineChunks(m2, m1)
	for _, u := range urls {
		fullUrl := baseURL + u
		resp, err := bc.Client.Get(fullUrl, false, nil)
		if err == nil {
			if bStr, ok := resp.Body.(string); ok {
				bc.RawHashes += bStr
			}
		}
	}

	return nil
}
