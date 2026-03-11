package spotapi

import (
	"fmt"
	"net/url"

	"github.com/spotapi/spotapi-go/internal/errors"
	"github.com/spotapi/spotapi-go/internal/types"
	"github.com/spotapi/spotapi-go/internal/utils"
)

type Login struct {
	Config                *types.Config
	Password              string
	IdentifierCredentials string
	Authorized            bool
	CsrfToken             string
	FlowId                string
}

func NewLogin(cfg *types.Config, password string, identifier string) *Login {
	l := &Login{
		Config:                cfg,
		Password:              password,
		IdentifierCredentials: identifier,
	}
	cfg.Client.FailException = func(msg, err string) error { return errors.NewLoginError(msg, err) }
	return l
}

func (l *Login) GetSession() error {
	loginUrl := "https://accounts.spotify.com/en/login"
	resp, err := l.Config.Client.Get(loginUrl, false, nil)
	if err != nil {
		return errors.NewLoginError("Could not get session", err.Error())
	}

	for _, cookie := range resp.Raw.Cookies() {
		if cookie.Name == "sp_sso_csrf_token" {
			l.CsrfToken = cookie.Value
			break
		}
	}

	bodyStr := fmt.Sprintf("%v", resp.Body)
	l.FlowId, _ = utils.ParseJsonString(bodyStr, "flowCtx")

	return nil
}

func (l *Login) Login() error {
	if l.Authorized {
		return errors.NewLoginError("User already logged in", "")
	}

	err := l.GetSession()
	if err != nil {
		return err
	}

	if l.Config.Solver == nil {
		return errors.NewLoginError("Solver not set", "")
	}

	captchaToken, err := l.Config.Solver.SolveCaptcha("https://accounts.spotify.com", "6LfCVLAUAAAAALFwwRnnCJ12DalriUGbj8FW_J39", "accounts/login", "v3")
	if err != nil {
		return err
	}

	return l.SubmitPassword(captchaToken)
}

func (l *Login) SubmitPassword(token string) error {
	data := url.Values{}
	data.Set("username", l.IdentifierCredentials)
	data.Set("password", l.Password)
	data.Set("recaptchaToken", token)
	data.Set("continue", fmt.Sprintf("https://open.spotify.com/?flow_ctx=%s", l.FlowId))
	data.Set("flowCtx", l.FlowId)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"X-Csrf-Token": l.CsrfToken,
	}

	resp, err := l.Config.Client.Post("https://accounts.spotify.com/login/password", false, headers, data.Encode())
	if err != nil {
		return err
	}

	if dataMap, ok := resp.Body.(map[string]interface{}); ok {
		if dataMap["result"] == "ok" {
			l.Authorized = true
			return nil
		}
	}

	return errors.NewLoginError("Login failed", fmt.Sprintf("%v", resp.Body))
}
