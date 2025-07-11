package ctfdapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

var nonceRegex = regexp.MustCompile(`<input id="nonce" name="nonce" type="hidden" value="([^"]+)">`)

// getNonce executes a GET request on the path endpoint and extracts the nonce from the hidden field of the HTML.
// The nonce is required for sending a POST request. Otherwise, the website will reject the request.
func (c *Client) getNonce(ctx context.Context, path string) (string, error) {
	return c.getByRegex(ctx, path, nonceRegex)
}

var csrfNonceRegex = regexp.MustCompile(`'csrfNonce': "([^"]+)",`)

func (c *Client) getCSRFNonce(ctx context.Context, path string) (string, error) {
	return c.getByRegex(ctx, path, csrfNonceRegex)
}

func (c *Client) getByRegex(ctx context.Context, path string, regex *regexp.Regexp) (string, error) {
	targetUrl, err := c.getTargetUrl(path, nil)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetUrl, nil)
	if err != nil {
		return "", fmt.Errorf("creating new HTTP request: %w", err)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("executing HTTP request: %w", err)
	}
	defer response.Body.Close() //nolint:errcheck

	pageData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d: %s", response.StatusCode, response.Status)
	}

	matches := regex.FindSubmatch(pageData)
	if len(matches) != 2 {
		return "", errors.New("CSRF nonce not found in HTML")
	}
	return string(matches[1]), nil
}

func (c *Client) sendGetRequest(ctx context.Context, path string, queryParameter map[string]string) ([]byte, error) {
	request, err := c.prepareRequest(ctx, http.MethodGet, path, queryParameter, nil)
	if err != nil {
		return nil, err
	}
	return c.executeRequest(request)
}

func (c *Client) sendPostRequest(ctx context.Context, path string, payload any) ([]byte, error) {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload into JSON: %w", err)
	}

	request, err := c.prepareRequest(ctx, http.MethodPost, path, nil, payloadData)
	if err != nil {
		return nil, err
	}
	return c.executeRequest(request)
}

func (c *Client) sendPatchRequest(ctx context.Context, path string, payload any) ([]byte, error) {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload into JSON: %w", err)
	}

	request, err := c.prepareRequest(ctx, http.MethodPatch, path, nil, payloadData)
	if err != nil {
		return nil, err
	}
	return c.executeRequest(request)
}

func (c *Client) sendDeleteRequest(ctx context.Context, path string) ([]byte, error) {
	request, err := c.prepareRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return c.executeRequest(request)
}

func (c *Client) prepareRequest(ctx context.Context, method string, path string, queryParameter map[string]string, body []byte) (*http.Request, error) {
	targetUrl, err := c.getTargetUrl(path, queryParameter)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	request, err := http.NewRequestWithContext(ctx, method, targetUrl, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating new HTTP request: %w", err)
	}
	if len(c.accessToken) != 0 {
		request.Header.Set("Authorization", "Token "+c.accessToken)
	}
	if len(c.csrfToken) != 0 {
		request.Header.Set("Csrf-Token", c.csrfToken)
	}
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

func (c *Client) executeRequest(request *http.Request) ([]byte, error) {
	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("executing HTTP request: %w", err)
	}
	defer response.Body.Close() //nolint:errcheck

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", response.StatusCode, response.Status)
	}
	return responseData, nil
}
