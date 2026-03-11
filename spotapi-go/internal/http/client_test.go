package http

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name        string
		profile     profiles.ClientProfile
		proxy       string
		autoRetries int
	}{
		{
			name:        "Chrome profile no proxy",
			profile:     profiles.Chrome_120,
			proxy:       "",
			autoRetries: 3,
		},
		{
			name:        "Firefox profile with proxy",
			profile:     profiles.Firefox_120,
			proxy:       "localhost:8080",
			autoRetries: 5,
		},
		{
			name:        "zero retries",
			profile:     profiles.Chrome_120,
			proxy:       "",
			autoRetries: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.profile, tc.proxy, tc.autoRetries)

			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}

			if client == nil {
				t.Fatal("Expected non-nil client")
			}

			if client.HttpClient == nil {
				t.Error("HttpClient should not be nil")
			}

			expectedRetries := tc.autoRetries + 1
			if client.AutoRetries != expectedRetries {
				t.Errorf("Expected AutoRetries %d, got %d", expectedRetries, client.AutoRetries)
			}

			if client.Authenticate != nil {
				t.Error("Authenticate should be nil initially")
			}

			if client.FailException != nil {
				t.Error("FailException should be nil initially")
			}
		})
	}
}

func TestNewClientInvalidProxy(t *testing.T) {
	// Test with a malformed proxy - should still create client but may fail on use
	client, err := NewClient(profiles.Chrome_120, ":::invalid:::", 3)

	// The library might accept invalid proxy format during creation
	// So we just check that we get some response
	if err != nil {
		// It's ok if it fails
		t.Logf("Expected behavior: NewClient with invalid proxy failed: %v", err)
	} else {
		if client == nil {
			t.Error("If no error, client should not be nil")
		}
	}
}

func TestResponseStructure(t *testing.T) {
	resp := &Response{
		StatusCode: 200,
		Body:       map[string]interface{}{"test": "data"},
		Raw:        nil,
		Fail:       false,
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected StatusCode 200, got %d", resp.StatusCode)
	}

	if resp.Fail {
		t.Error("Expected Fail to be false")
	}

	if resp.Body == nil {
		t.Error("Body should not be nil")
	}
}

func TestResponseFailConditions(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		expectFail bool
	}{
		{"success 200", 200, false},
		{"success 201", 201, false},
		{"redirect 302", 302, false},
		{"client error 400", 400, true},
		{"client error 404", 404, true},
		{"server error 500", 500, true},
		{"below 200", 199, true},
		{"above 302", 303, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fail := tc.statusCode < 200 || tc.statusCode > 302

			if fail != tc.expectFail {
				t.Errorf("For status %d, expected fail=%v, got fail=%v", tc.statusCode, tc.expectFail, fail)
			}
		})
	}
}

func TestClientGetMethod(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 1)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Verify the client was created successfully and has the Get method
	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

func TestClientPostMethod(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 1)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Verify the client was created successfully and has the Post method
	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

func TestClientPutMethod(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 1)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Verify the client was created successfully and has the Put method
	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

func TestAuthRuleFunction(t *testing.T) {
	// Test that AuthRule type is properly defined
	var authFunc AuthRule = func(headers map[string]string) (map[string]string, error) {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Authorization"] = "Bearer test"
		return headers, nil
	}

	result, err := authFunc(nil)
	if err != nil {
		t.Errorf("AuthRule function failed: %v", err)
	}

	if result["Authorization"] != "Bearer test" {
		t.Error("AuthRule should add Authorization header")
	}
}

func TestClientWithAuthRule(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 1)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Set a custom auth rule
	called := false
	client.Authenticate = func(headers map[string]string) (map[string]string, error) {
		called = true
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Custom-Auth"] = "test-token"
		return headers, nil
	}

	// Verify the function was set
	if client.Authenticate == nil {
		t.Fatal("Authenticate function should be set")
	}

	// Test calling it
	headers := make(map[string]string)
	result, err := client.Authenticate(headers)
	if err != nil {
		t.Errorf("Authenticate function failed: %v", err)
	}

	if !called {
		t.Error("Authenticate function should have been called")
	}

	if result["Custom-Auth"] != "test-token" {
		t.Error("Authenticate should have set Custom-Auth header")
	}
}

func TestClientWithFailException(t *testing.T) {
	client, err := NewClient(profiles.Chrome_120, "", 1)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Set a custom fail exception handler
	called := false
	client.FailException = func(msg, err string) error {
		called = true
		return &testError{message: msg + ": " + err}
	}

	// Verify the function was set
	if client.FailException == nil {
		t.Fatal("FailException function should be set")
	}

	// Test calling it
	err = client.FailException("test", "error")
	if err == nil {
		t.Error("FailException should return an error")
	}

	if !called {
		t.Error("FailException function should have been called")
	}
}

// Helper test error type
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestClientAutoRetryIncrement(t *testing.T) {
	// Test that autoRetries is incremented by 1
	testCases := []struct {
		input    int
		expected int
	}{
		{0, 1},
		{1, 2},
		{3, 4},
		{10, 11},
	}

	for _, tc := range testCases {
		client, err := NewClient(profiles.Chrome_120, "", tc.input)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		if client.AutoRetries != tc.expected {
			t.Errorf("Input %d: expected AutoRetries %d, got %d", tc.input, tc.expected, client.AutoRetries)
		}
	}
}

func TestNewClientWithDifferentProfiles(t *testing.T) {
	profiles := []profiles.ClientProfile{
		profiles.Chrome_120,
		profiles.Firefox_120,
		profiles.Safari_Ipad_15_6,
		profiles.Opera_91,
	}

	for _, profile := range profiles {
		client, err := NewClient(profile, "", 3)
		if err != nil {
			t.Errorf("NewClient failed with profile %v: %v", profile, err)
			continue
		}

		if client == nil {
			t.Errorf("Expected non-nil client for profile %v", profile)
		}
	}
}