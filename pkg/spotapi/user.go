package spotapi

import (
	"fmt"

	"github.com/spotapi/spotapi-go/internal/errors"
)

type User struct {
	Login     *Login
	userPlan  map[string]interface{}
	userInfo  map[string]interface{}
	csrfToken string
}

func NewUser(l *Login) (*User, error) {
	if !l.Authorized {
		return nil, fmt.Errorf("must be logged in")
	}

	return &User{
		Login: l,
	}, nil
}

func (u *User) HasPremium() (bool, error) {
	if u.userPlan == nil {
		plan, err := u.GetPlanInfo()
		if err != nil {
			return false, err
		}
		u.userPlan = plan
	}

	planData, ok := u.userPlan["plan"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid plan data")
	}

	name, ok := planData["name"].(string)
	if !ok {
		return false, fmt.Errorf("plan name is not a string")
	}

	return name != "Spotify Free", nil
}

func (u *User) Username() (string, error) {
	if u.userInfo == nil {
		info, err := u.GetUserInfo()
		if err != nil {
			return "", err
		}
		u.userInfo = info
	}

	profile, ok := u.userInfo["profile"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid user info")
	}

	username, ok := profile["username"].(string)
	if !ok {
		return "", fmt.Errorf("username is not a string")
	}

	return username, nil
}

func (u *User) GetPlanInfo() (map[string]interface{}, error) {
	url := "https://www.spotify.com/ca-en/api/account/v2/plan/"
	resp, err := u.Login.Config.Client.Get(url, false, nil)
	if err != nil {
		return nil, errors.NewUserError("Could not get user plan info", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, errors.NewUserError("Invalid JSON", "")
}

func (u *User) GetUserInfo() (map[string]interface{}, error) {
	url := "https://www.spotify.com/api/account-settings/v1/profile"
	resp, err := u.Login.Config.Client.Get(url, false, nil)
	if err != nil {
		return nil, errors.NewUserError("Could not get user info", err.Error())
	}

	if data, ok := resp.Body.(map[string]interface{}); ok {
		u.csrfToken = resp.Raw.Header.Get("X-Csrf-Token")
		return data, nil
	}

	return nil, errors.NewUserError("Invalid JSON", "")
}

func (u *User) EditUserInfo(dump map[string]interface{}) error {
	if u.Login.Config.Solver == nil {
		return errors.NewUserError("Captcha solver not set", "")
	}

	if u.csrfToken == "" {
		if _, err := u.GetUserInfo(); err != nil {
			return errors.NewUserError("Could not ensure CSRF token", err.Error())
		}
	}

	captchaResponse, err := u.Login.Config.Solver.SolveCaptcha(
		"https://www.spotify.com",
		"6LfCVLAUAAAAALFwwRnnCJ12DalriUGbj8FW_J39",
		"account_settings/profile_update",
		"v3",
	)
	if err != nil {
		return err
	}

	profileDump, ok := dump["profile"].(map[string]interface{})
	if !ok {
		return errors.NewUserError("Invalid profile dump", "")
	}

	newDump := map[string]interface{}{
		"profile": map[string]interface{}{
			"email":     profileDump["email"],
			"gender":    profileDump["gender"],
			"birthdate": profileDump["birthdate"],
			"country":   profileDump["country"],
		},
		"recaptcha_token": captchaResponse,
		"client_nonce":    "", // use utils.RandomNonce when implemented
		"callback_url":    "https://www.spotify.com/account/profile/challenge",
		"client_info":     map[string]interface{}{"locale": "en_US", "capabilities": []int{1}},
	}

	url := "https://www.spotify.com/api/account-settings/v2/profile"
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Csrf-Token": u.csrfToken,
	}

	_, err = u.Login.Config.Client.Put(url, false, headers, newDump)
	if err != nil {
		return errors.NewUserError("Could not edit user info", err.Error())
	}

	return nil
}
