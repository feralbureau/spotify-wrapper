package types

import (
	"github.com/spotapi/spotapi-go/internal/http"
)

type Logger interface {
	Info(s string, extra ...interface{})
	Attempt(s string, extra ...interface{})
	Error(s string, extra ...interface{})
	Fatal(s string, extra ...interface{})
}

type CaptchaSolver interface {
	SolveCaptcha(url, siteKey, action, task string) (string, error)
}

type Config struct {
	Logger Logger
	Solver CaptchaSolver
	Client *http.Client
}

type Saver interface {
	Save(data []map[string]interface{}) error
	Load(query map[string]interface{}) (map[string]interface{}, error)
	Delete(query map[string]interface{}) error
}
