package ctfdapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	loginPath = "/login"
)

type LoginRequest struct {
	Name     string
	Password string
}

func (c *Client) Login(ctx context.Context, loginRequest LoginRequest) error {
	nonce, err := c.getNonce(ctx, loginPath)
	if err != nil {
		return fmt.Errorf("getting nonce: %w", err)
	}
	if err := c.loginSendForm(ctx, loginRequest, nonce); err != nil {
		return fmt.Errorf("sending form: %w", err)
	}
	return nil
}

// loginSendForm constructs a POST request to the login endpoint with the configuration provided by the LoginRequest.
//
//nolint:dupl
func (c *Client) loginSendForm(ctx context.Context, loginRequest LoginRequest, nonce string) error {
	targetUrl, err := c.getTargetUrl(loginPath)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	multipartWriter := multipart.NewWriter(&body)
	if err := c.loginRequestToForm(multipartWriter, loginRequest); err != nil {
		return err
	}
	if err := multipartWriter.WriteField("_submit", "Submit"); err != nil {
		return err
	}
	if err := multipartWriter.WriteField("nonce", nonce); err != nil {
		return err
	}
	if err := multipartWriter.Close(); err != nil {
		return fmt.Errorf("closing multipart writer: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, targetUrl, &body)
	if err != nil {
		return fmt.Errorf("creating new HTTP request: %w", err)
	}
	request.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("executing HTTP request: %w", err)
	}
	defer response.Body.Close() //nolint:errcheck

	_, err = io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if response.StatusCode != http.StatusFound {
		return fmt.Errorf("unexpected status code %d: %s", response.StatusCode, response.Status)
	}
	return nil
}

func (c *Client) loginRequestToForm(writer *multipart.Writer, loginRequest LoginRequest) error {
	if err := writer.WriteField("name", loginRequest.Name); err != nil {
		return err
	}
	if err := writer.WriteField("password", loginRequest.Password); err != nil {
		return err
	}
	return nil
}
