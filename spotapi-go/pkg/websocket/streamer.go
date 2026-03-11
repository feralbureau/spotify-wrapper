package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bogdanfinn/websocket"
	"github.com/spotapi/spotapi-go/internal/http"
	"github.com/spotapi/spotapi-go/internal/utils"
	"github.com/spotapi/spotapi-go/pkg/spotapi"
)

type WebsocketStreamer struct {
	Base         *http.BaseClient
	DeviceId     string
	Conn         *websocket.Conn
	ConnectionId string
	mu           sync.Mutex
}

// NewWebsocketStreamer creates a WebsocketStreamer connected to Spotify's dealer websocket for the given login.
// 
// It validates that the login is authorized, initializes an HTTP base client and session tokens, generates a random
// device ID, establishes a websocket connection using the client's access token, reads the initial packet to obtain
// and set the connection ID, and starts a background keep-alive goroutine.
// 
// The returned WebsocketStreamer is ready for receiving packets via GetPacket.
// An error is returned if the login is not authorized, the websocket dial fails, or the initial packet cannot be parsed.
func NewWebsocketStreamer(l *spotapi.Login) (*WebsocketStreamer, error) {
	if !l.Authorized {
		return nil, fmt.Errorf("must be logged in")
	}

	bc := http.NewBaseClient(l.Config.Client, "en")
	bc.GetSession()
	bc.GetClientToken()

	deviceId := utils.RandomHexString(32)
	uri := fmt.Sprintf("wss://dealer.spotify.com/?access_token=%s", bc.AccessToken)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(uri, nil)
	if err != nil {
		return nil, err
	}

	s := &WebsocketStreamer{
		Base:     bc,
		DeviceId: deviceId,
		Conn:     conn,
	}

	s.ConnectionId, err = s.getInitPacket()
	if err != nil {
		return nil, err
	}

	go s.keepAlive()

	return s, nil
}

func (s *WebsocketStreamer) getInitPacket() (string, error) {
	_, message, err := s.Conn.ReadMessage()
	if err != nil {
		return "", err
	}

	var packet map[string]interface{}
	json.Unmarshal(message, &packet)

	headers, ok := packet["headers"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid init packet")
	}

	connId, ok := headers["Spotify-Connection-Id"].(string)
	if !ok {
		return "", fmt.Errorf("no connection id")
	}

	return connId, nil
}

func (s *WebsocketStreamer) keepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		s.Conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
		s.mu.Unlock()
	}
}

func (s *WebsocketStreamer) GetPacket() (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, message, err := s.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	var packet map[string]interface{}
	json.Unmarshal(message, &packet)
	return packet, nil
}
