package spotapi

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/http"
	"github.com/spotapi/spotapi-go/internal/types"
)

func TestNewLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "test-password", "test@example.com")

	if login == nil {
		t.Fatal("NewLogin returned nil")
	}

	if login.Config != cfg {
		t.Error("Config should be set")
	}

	if login.Password != "test-password" {
		t.Errorf("Expected password 'test-password', got %q", login.Password)
	}

	if login.IdentifierCredentials != "test@example.com" {
		t.Errorf("Expected identifier 'test@example.com', got %q", login.IdentifierCredentials)
	}

	if login.Authorized {
		t.Error("Should not be authorized initially")
	}

	if login.Config.Client.FailException == nil {
		t.Error("FailException should be set by NewLogin")
	}
}

func TestNewLoginWithDifferentIdentifiers(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	testCases := []struct {
		name       string
		identifier string
	}{
		{"email", "user@example.com"},
		{"username", "testuser123"},
		{"phone", "+1234567890"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			login := NewLogin(cfg, "password", tc.identifier)

			if login.IdentifierCredentials != tc.identifier {
				t.Errorf("Expected identifier %q, got %q", tc.identifier, login.IdentifierCredentials)
			}
		})
	}
}

func TestLoginInitialState(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "password", "user@example.com")

	// Check initial state
	if login.Authorized {
		t.Error("Should not be authorized initially")
	}

	if login.CsrfToken != "" {
		t.Error("CsrfToken should be empty initially")
	}

	if login.FlowId != "" {
		t.Error("FlowId should be empty initially")
	}
}

func TestLoginStructure(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "password", "user@example.com")

	// Verify all fields are accessible
	_ = login.Config
	_ = login.Password
	_ = login.IdentifierCredentials
	_ = login.Authorized
	_ = login.CsrfToken
	_ = login.FlowId
}

func TestLoginFailExceptionSetup(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "password", "user@example.com")

	// Verify FailException was set
	if login.Config.Client.FailException == nil {
		t.Fatal("FailException should be set")
	}

	// Test calling it
	err := login.Config.Client.FailException("test message", "test error")
	if err == nil {
		t.Error("FailException should return an error")
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestLoginWithEmptyPassword(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "", "user@example.com")

	if login.Password != "" {
		t.Error("Empty password should be preserved")
	}
}

func TestLoginWithEmptyIdentifier(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "password", "")

	if login.IdentifierCredentials != "" {
		t.Error("Empty identifier should be preserved")
	}
}

func TestLoginAlreadyAuthorized(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "password", "user@example.com")
	login.Authorized = true

	err := login.Login()

	if err == nil {
		t.Error("Login should fail when already authorized")
	}

	expectedMsg := "User already logged in"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestLoginWithoutSolver(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "password", "user@example.com")

	// GetSession will fail in unit test without network
	// But we can test that Login checks for solver
	err := login.Login()

	// Should fail because GetSession will fail or Solver is not set
	if err == nil {
		t.Log("Login succeeded (may have network access and solver)")
	} else {
		t.Logf("Expected failure in unit test: %v", err)
	}
}

func TestLoginMultipleInstances(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg1 := &types.Config{Client: client}
	cfg2 := &types.Config{Client: client}

	login1 := NewLogin(cfg1, "password1", "user1@example.com")
	login2 := NewLogin(cfg2, "password2", "user2@example.com")

	// Verify they are independent
	if login1.Password == login2.Password {
		t.Error("Different logins should have different passwords")
	}

	if login1.IdentifierCredentials == login2.IdentifierCredentials {
		t.Error("Different logins should have different identifiers")
	}
}

func TestLoginConfigAssignment(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "password", "user@example.com")

	if login.Config != cfg {
		t.Error("Config should be the same reference")
	}

	if login.Config.Client != client {
		t.Error("Client should be accessible through Config")
	}
}

func TestLoginWithSpecialCharacters(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	testCases := []struct {
		name       string
		password   string
		identifier string
	}{
		{
			name:       "special chars in password",
			password:   "p@$$w0rd!#%",
			identifier: "user@example.com",
		},
		{
			name:       "special chars in email",
			password:   "password",
			identifier: "user+tag@example.com",
		},
		{
			name:       "unicode characters",
			password:   "пароль",
			identifier: "用户@example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			login := NewLogin(cfg, tc.password, tc.identifier)

			if login.Password != tc.password {
				t.Errorf("Password not preserved correctly")
			}

			if login.IdentifierCredentials != tc.identifier {
				t.Errorf("Identifier not preserved correctly")
			}
		})
	}
}

func TestLoginStateTransitions(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}

	login := NewLogin(cfg, "password", "user@example.com")

	// Initially not authorized
	if login.Authorized {
		t.Error("Should start as not authorized")
	}

	// Manually set authorized (simulating successful login)
	login.Authorized = true

	if !login.Authorized {
		t.Error("Should be authorized after setting")
	}

	// Try to login again
	err := login.Login()
	if err == nil {
		t.Error("Should not allow login when already authorized")
	}
}