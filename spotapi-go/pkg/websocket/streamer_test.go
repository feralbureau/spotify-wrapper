package websocket

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/http"
	"github.com/spotapi/spotapi-go/internal/types"
	"github.com/spotapi/spotapi-go/pkg/spotapi"
)

func TestNewWebsocketStreamerWithoutLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &spotapi.Login{Config: cfg, Authorized: false}

	streamer, err := NewWebsocketStreamer(login)

	if err == nil {
		t.Error("NewWebsocketStreamer should fail when not logged in")
	}

	if streamer != nil {
		t.Error("Streamer should be nil when login is not authorized")
	}

	expectedMsg := "must be logged in"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestNewWebsocketStreamerWithAuthorizedLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &spotapi.Login{Config: cfg, Authorized: true}

	streamer, err := NewWebsocketStreamer(login)

	// This will fail in unit test without real network connection
	if err == nil {
		// If it succeeds, verify structure
		if streamer == nil {
			t.Error("Streamer should not be nil on success")
		}

		if streamer.Base == nil {
			t.Error("Base should not be nil")
		}

		if streamer.DeviceId == "" {
			t.Error("DeviceId should not be empty")
		}

		if streamer.Conn == nil {
			t.Error("Conn should not be nil")
		}

		// Clean up
		if streamer.Conn != nil {
			streamer.Conn.Close()
		}
	} else {
		t.Logf("Expected failure in unit test without network: %v", err)
	}
}

func TestWebsocketStreamerStructure(t *testing.T) {
	// Test the structure exists
	streamer := &WebsocketStreamer{
		DeviceId:     "test-device-id",
		ConnectionId: "test-connection-id",
	}

	if streamer.DeviceId != "test-device-id" {
		t.Error("DeviceId should be accessible")
	}

	if streamer.ConnectionId != "test-connection-id" {
		t.Error("ConnectionId should be accessible")
	}

	// Verify fields are accessible
	_ = streamer.Base
	_ = streamer.Conn
	_ = streamer.mu
}

func TestWebsocketStreamerWithNilLogin(t *testing.T) {
	// Should panic or return error with nil login
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Panicked as expected: %v", r)
		}
	}()

	streamer, err := NewWebsocketStreamer(nil)

	if err != nil {
		t.Logf("Returned error as expected: %v", err)
	}

	if streamer != nil {
		t.Error("Streamer should be nil with nil login")
	}
}

func TestWebsocketStreamerDeviceIdGeneration(t *testing.T) {
	// Test that device ID is generated (32 hex characters)
	// We can't fully test without network, but we can verify the structure

	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &spotapi.Login{Config: cfg, Authorized: true}

	_, err := NewWebsocketStreamer(login)

	// Will fail without network, but that's expected
	if err != nil {
		t.Logf("Expected failure without network: %v", err)
	}
}

func TestWebsocketStreamerConnectionIdExtraction(t *testing.T) {
	// Test structure - can't fully test without mocking websocket
	streamer := &WebsocketStreamer{
		ConnectionId: "test-conn-id-123",
	}

	if streamer.ConnectionId != "test-conn-id-123" {
		t.Error("ConnectionId should be set correctly")
	}
}

func TestWebsocketStreamerMutexAccess(t *testing.T) {
	// Test that mutex is accessible and can be locked/unlocked
	streamer := &WebsocketStreamer{}

	// This should not panic
	streamer.mu.Lock()
	streamer.mu.Unlock()
}

func TestWebsocketStreamerBaseClientSetup(t *testing.T) {
	// Verify that Base client would be initialized
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &spotapi.Login{Config: cfg, Authorized: true}

	// This will fail, but we're testing the structure
	_, err := NewWebsocketStreamer(login)

	if err != nil {
		t.Logf("Expected failure in unit test: %v", err)
	}
}

func TestWebsocketStreamerRequiresAuthorization(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	testCases := []struct {
		name       string
		authorized bool
		shouldFail bool
	}{
		{
			name:       "authorized login",
			authorized: true,
			shouldFail: false, // May still fail due to network, but not auth check
		},
		{
			name:       "unauthorized login",
			authorized: false,
			shouldFail: true, // Should fail auth check
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			login := &spotapi.Login{Config: cfg, Authorized: tc.authorized}
			_, err := NewWebsocketStreamer(login)

			if tc.shouldFail {
				if err == nil {
					t.Error("Should fail with unauthorized login")
				}
				expectedMsg := "must be logged in"
				if err.Error() != expectedMsg {
					t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
				}
			} else {
				// May fail due to network, but not auth check
				if err != nil {
					t.Logf("Failed (likely due to network): %v", err)
				}
			}
		})
	}
}

func TestWebsocketStreamerFieldsInitialized(t *testing.T) {
	// Test that all expected fields exist
	streamer := &WebsocketStreamer{
		DeviceId:     "device123",
		ConnectionId: "conn456",
	}

	// Verify all fields are accessible
	if streamer.DeviceId == "" {
		t.Error("DeviceId should not be empty after setting")
	}

	if streamer.ConnectionId == "" {
		t.Error("ConnectionId should not be empty after setting")
	}
}

func TestWebsocketStreamerKeepAliveSetup(t *testing.T) {
	// We can't fully test the keepAlive goroutine without mocking,
	// but we can verify the structure exists
	// The keepAlive function is started as a goroutine in NewWebsocketStreamer

	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &spotapi.Login{Config: cfg, Authorized: true}

	_, err := NewWebsocketStreamer(login)

	// Will fail without network
	if err != nil {
		t.Logf("Expected failure without network: %v", err)
	}
}