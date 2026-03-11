package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spotapi/spotapi-go/internal/errors"
	"github.com/spotapi/spotapi-go/internal/totp"
)

type BaseClient struct {
	Client        *Client
	JsPack        string
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
	headers["Accept-Language"] = bc.Language

	return headers, nil
}

func (bc *BaseClient) GetSession() error {
	resp, err := bc.Client.Get("https://open.spotify.com", false, nil)
	if err != nil {
		return errors.NewBaseClientError("Could not get session", err.Error())
	}

	bodyStr, ok := resp.Body.(string)
	if !ok {
		return errors.NewBaseClientError("Could not get session", "Response body is not a string")
	}

	// Extract appServerConfig
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

		if v, ok := bc.ServerCfg["clientVersion"].(string); ok {
			bc.ClientVersion = v
		}
	}

	// Device ID from cookie
	for _, cookie := range resp.Raw.Cookies() {
		if cookie.Name == "sp_t" {
			bc.DeviceId = cookie.Value
			break
		}
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
			if !ok1 || !ok2 {
				return errors.NewBaseClientError("Could not get session auth tokens", "Invalid response format")
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

	resp, err := bc.Client.Post(url, false, nil, payload)
	if err != nil {
		return errors.NewBaseClientError("Could not get client token", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		if gt, ok := data["granted_token"].(map[string]interface{}); ok {
			token, ok := gt["token"].(string)
			if !ok {
				return errors.NewBaseClientError("Could not get client token", "Token is not a string")
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

func (bc *BaseClient) PartHash(name string) string {
	if bc.RawHashes == "" {
		bc.GetSha256Hash()
	}

	// Simplified hash extraction
	searchQuery := fmt.Sprintf("\"%s\",\"query\",\"", name)
	if idx := strings.Index(bc.RawHashes, searchQuery); idx != -1 {
		start := idx + len(searchQuery)
		end := strings.Index(bc.RawHashes[start:], "\"")
		return bc.RawHashes[start : start+end]
	}

	searchMutation := fmt.Sprintf("\"%s\",\"mutation\",\"", name)
	if idx := strings.Index(bc.RawHashes, searchMutation); idx != -1 {
		start := idx + len(searchMutation)
		end := strings.Index(bc.RawHashes[start:], "\"")
		return bc.RawHashes[start : start+end]
	}

	return ""
}

func (bc *BaseClient) GetSha256Hash() error {
	// Logic to fetch and parse hashes from JS packs
	return nil
}
