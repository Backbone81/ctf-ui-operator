package ctfdapi

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

type Client struct {
	baseUrl     *url.URL
	client      *http.Client
	accessToken string
	csrfToken   string
}

func NewClient(baseUrl string, accessToken string) (*Client, error) {
	parsedBaseUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}

	// The client needs to work with cookies. Otherwise, endpoints like /setup will not work.
	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("creating a cookie jar: %w", err)
	}

	httpClient := &http.Client{
		Jar: cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// We do not want to automatically follow redirects, because this would make it difficult to detect if
			// the instance is already setup or not. The /setup endpoint returns an HTTP 302 when the setup was already
			// done. This would be swallowed if we followed redirects automatically.
			return http.ErrUseLastResponse
		},
	}
	return &Client{
		baseUrl:     parsedBaseUrl,
		client:      httpClient,
		accessToken: accessToken,
	}, nil
}

func (c *Client) getTargetUrl(urlPath string, queryParameter map[string]string) (string, error) {
	path, err := url.JoinPath(c.baseUrl.String(), urlPath)
	if err != nil {
		return "", err
	}
	params := url.Values{}
	for key, value := range queryParameter {
		params.Add(key, value)
	}
	return path + "?" + params.Encode(), nil
}
