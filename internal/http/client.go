package http

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/spotapi/spotapi-go/internal/errors"
)

type Response struct {
	StatusCode int
	Body       interface{}
	Raw        *fhttp.Response
	Fail       bool
}

type AuthRule func(headers map[string]string) (map[string]string, error)

type Client struct {
	HttpClient    tls_client.HttpClient
	AutoRetries   int
	Authenticate  AuthRule
	FailException func(string, string) error
}

func NewClient(profile profiles.ClientProfile, proxy string, autoRetries int) (*Client, error) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profile),
		tls_client.WithRandomTLSExtensionOrder(),
	}

	if proxy != "" {
		options = append(options, tls_client.WithProxyUrl("http://"+proxy))
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, err
	}

	return &Client{
		HttpClient:  client,
		AutoRetries: autoRetries + 1,
	}, nil
}

func (c *Client) Request(method, url string, authenticate bool, headers map[string]string, body interface{}) (*Response, error) {
	if authenticate && c.Authenticate != nil {
		var err error
		headers, err = c.Authenticate(headers)
		if err != nil {
			return nil, err
		}
	}

	var reqBody string
	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = v
		case []byte:
			reqBody = string(v)
		default:
			b, _ := json.Marshal(v)
			reqBody = string(b)
		}
	}

	var lastErr error
	for i := 0; i < c.AutoRetries; i++ {
		req, err := fhttp.NewRequest(method, url, strings.NewReader(reqBody))
		if err != nil {
			lastErr = err
			continue
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.HttpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		parsedResp := c.parseResponse(resp)

		if parsedResp.Fail && c.FailException != nil {
			return parsedResp, c.FailException(fmt.Sprintf("Could not %s %s. Status Code: %d", method, url, parsedResp.StatusCode), "Request Failed.")
		}

		return parsedResp, nil
	}

	errMsg := "Unknown"
	if lastErr != nil {
		errMsg = lastErr.Error()
	}
	return nil, errors.NewRequestError("Failed to complete request.", errMsg)
}

func (c *Client) parseResponse(resp *fhttp.Response) *Response {
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)

	var body interface{}
	err := json.Unmarshal(bodyBytes, &body)
	if err != nil {
		body = string(bodyBytes)
	}

	fail := resp.StatusCode < 200 || resp.StatusCode > 302

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Raw:        resp,
		Fail:       fail,
	}
}

func (c *Client) Get(url string, authenticate bool, headers map[string]string) (*Response, error) {
	return c.Request("GET", url, authenticate, headers, nil)
}

func (c *Client) Post(url string, authenticate bool, headers map[string]string, body interface{}) (*Response, error) {
	return c.Request("POST", url, authenticate, headers, body)
}

func (c *Client) Put(url string, authenticate bool, headers map[string]string, body interface{}) (*Response, error) {
	return c.Request("PUT", url, authenticate, headers, body)
}
