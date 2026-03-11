package spotapi

import (
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/http"
	"github.com/spotapi/spotapi-go/internal/types"
)

func TestNewUserWithAuthorizedLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, err := NewUser(login)

	if err != nil {
		t.Fatalf("NewUser failed: %v", err)
	}

	if user == nil {
		t.Fatal("NewUser returned nil")
	}

	if user.Login != login {
		t.Error("Login should be set")
	}
}

func TestNewUserWithUnauthorizedLogin(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: false}

	user, err := NewUser(login)

	if err == nil {
		t.Error("NewUser should fail with unauthorized login")
	}

	if user != nil {
		t.Error("NewUser should return nil when login is not authorized")
	}

	expectedMsg := "must be logged in"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestUserInitialState(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Check initial state
	if user.userPlan != nil {
		t.Error("userPlan should be nil initially")
	}

	if user.userInfo != nil {
		t.Error("userInfo should be nil initially")
	}

	if user.csrfToken != "" {
		t.Error("csrfToken should be empty initially")
	}
}

func TestUserStructure(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Verify all fields are accessible
	_ = user.Login
	_ = user.userPlan
	_ = user.userInfo
	_ = user.csrfToken
}

func TestHasPremiumWithoutCachedPlan(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Will fail because we can't make real API calls
	_, err := user.HasPremium()

	if err == nil {
		t.Log("HasPremium succeeded (may have network access)")
	} else {
		t.Logf("Expected failure in unit test without network: %v", err)
	}
}

func TestHasPremiumWithCachedPlan(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Manually set cached plan data
	user.userPlan = map[string]interface{}{
		"plan": map[string]interface{}{
			"name": "Premium",
		},
	}

	hasPremium, err := user.HasPremium()

	if err != nil {
		t.Errorf("HasPremium failed: %v", err)
	}

	if !hasPremium {
		t.Error("User with Premium plan should return true")
	}
}

func TestHasPremiumWithFreePlan(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Manually set cached plan data for free user
	user.userPlan = map[string]interface{}{
		"plan": map[string]interface{}{
			"name": "Spotify Free",
		},
	}

	hasPremium, err := user.HasPremium()

	if err != nil {
		t.Errorf("HasPremium failed: %v", err)
	}

	if hasPremium {
		t.Error("User with Spotify Free should return false")
	}
}

func TestUsernameWithoutCachedInfo(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Will fail because we can't make real API calls
	_, err := user.Username()

	if err == nil {
		t.Log("Username succeeded (may have network access)")
	} else {
		t.Logf("Expected failure in unit test without network: %v", err)
	}
}

func TestUsernameWithCachedInfo(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Manually set cached user info
	user.userInfo = map[string]interface{}{
		"profile": map[string]interface{}{
			"username": "testuser123",
		},
	}

	username, err := user.Username()

	if err != nil {
		t.Errorf("Username failed: %v", err)
	}

	if username != "testuser123" {
		t.Errorf("Expected username 'testuser123', got %q", username)
	}
}

func TestEditUserInfoWithoutSolver(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	dump := map[string]interface{}{
		"profile": map[string]interface{}{
			"email":     "new@example.com",
			"gender":    "male",
			"birthdate": "1990-01-01",
			"country":   "US",
		},
	}

	err := user.EditUserInfo(dump)

	if err == nil {
		t.Error("EditUserInfo should fail when solver is not set")
	}

	expectedMsg := "Captcha solver not set"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestEditUserInfoWithInvalidDump(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Invalid dump without "profile" key
	dump := map[string]interface{}{
		"invalid": "data",
	}

	err := user.EditUserInfo(dump)

	if err == nil {
		t.Error("EditUserInfo should fail with invalid dump")
	}
}

func TestUserMultipleInstances(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg1 := &types.Config{Client: client}
	cfg2 := &types.Config{Client: client}
	login1 := &Login{Config: cfg1, Authorized: true}
	login2 := &Login{Config: cfg2, Authorized: true}

	user1, _ := NewUser(login1)
	user2, _ := NewUser(login2)

	// Verify they are independent
	if user1 == user2 {
		t.Error("Different user instances should be different objects")
	}

	if user1.Login == user2.Login {
		t.Error("Different users should have different Login instances")
	}
}

func TestUserPlanCaching(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Set cached plan
	user.userPlan = map[string]interface{}{
		"plan": map[string]interface{}{
			"name": "Premium",
		},
	}

	// First call should use cached data
	result1, _ := user.HasPremium()

	// Modify cached data
	user.userPlan = map[string]interface{}{
		"plan": map[string]interface{}{
			"name": "Spotify Free",
		},
	}

	// Second call should use new cached data
	result2, _ := user.HasPremium()

	if result1 == result2 {
		t.Error("Results should differ after cache modification")
	}
}

func TestUserInfoCaching(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Set cached info
	user.userInfo = map[string]interface{}{
		"profile": map[string]interface{}{
			"username": "user1",
		},
	}

	// First call should use cached data
	name1, _ := user.Username()

	// Modify cached data
	user.userInfo = map[string]interface{}{
		"profile": map[string]interface{}{
			"username": "user2",
		},
	}

	// Second call should use new cached data
	name2, _ := user.Username()

	if name1 == name2 {
		t.Error("Results should differ after cache modification")
	}

	if name1 != "user1" || name2 != "user2" {
		t.Error("Username should reflect cached values")
	}
}

func TestHasPremiumInvalidPlanData(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Set invalid plan data (missing "plan" key)
	user.userPlan = map[string]interface{}{
		"invalid": "data",
	}

	_, err := user.HasPremium()

	if err == nil {
		t.Error("HasPremium should fail with invalid plan data")
	}
}

func TestUsernameInvalidUserInfo(t *testing.T) {
	client, _ := http.NewClient(profiles.Chrome_120, "", 3)
	cfg := &types.Config{Client: client}
	login := &Login{Config: cfg, Authorized: true}

	user, _ := NewUser(login)

	// Set invalid user info (missing "profile" key)
	user.userInfo = map[string]interface{}{
		"invalid": "data",
	}

	_, err := user.Username()

	if err == nil {
		t.Error("Username should fail with invalid user info")
	}
}