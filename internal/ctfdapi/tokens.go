package ctfdapi

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"strconv"
	"time"
)

const (
	tokensPath = "/api/v1/tokens" //nolint:gosec // This is no hardcoded credential.
)

type Token struct {
	Id          int       `json:"id"`
	Type        string    `json:"type"`
	UserId      int       `json:"user_id"` //nolint:tagliatelle // we cannot change the spelling
	Created     time.Time `json:"created"`
	Expiration  time.Time `json:"expiration"`
	Description string    `json:"description"`
	Value       string    `json:"value"`
}

type ListTokensResponse struct {
	Success bool    `json:"success"`
	Data    []Token `json:"data"`
}

func (c *Client) ListTokens(ctx context.Context) (ListTokensResponse, error) {
	data, err := c.sendGetRequest(ctx, tokensPath, nil)
	if err != nil {
		return ListTokensResponse{}, err
	}

	var response ListTokensResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return ListTokensResponse{}, err
	}

	if !response.Success {
		return response, errors.New("the API request did not succeed")
	}
	return response, nil
}

type CreateTokenRequest struct {
	Description string   `json:"description"`
	Expiration  DateOnly `json:"expiration,omitempty"`
}

type CreateTokenResponse struct {
	Success bool  `json:"success"`
	Data    Token `json:"data"`
}

// CreateToken creates a new access token.
func (c *Client) CreateToken(ctx context.Context, createTokenRequest CreateTokenRequest) (CreateTokenResponse, error) {
	if len(c.accessToken) == 0 {
		// When this function is called while we do not yet have an access token at hand, we need to simulate a web
		// browser navigating the web UI and triggering API endpoints. To be able to call an API endpoint without
		// an access token, we need a session cookie active, which can be set on the client through a call to Login()
		// and we need to fetch a CSRF token from the page which is usually displayed before calling the API endpoint.
		// We provide that CSRF token for the next api call and make sure to clear it again when we are done.
		csrfNonce, err := c.getCSRFNonce(ctx, "/settings")
		if err != nil {
			return CreateTokenResponse{}, err
		}
		c.csrfToken = csrfNonce
		defer func() { c.csrfToken = "" }()
	}

	data, err := c.sendPostRequest(ctx, tokensPath, createTokenRequest)
	if err != nil {
		return CreateTokenResponse{}, err
	}

	var response CreateTokenResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return CreateTokenResponse{}, err
	}

	if !response.Success {
		return response, errors.New("the API request did not succeed")
	}
	return response, nil
}

type DeleteTokenResponse struct {
	Success bool `json:"success"`
}

func (c *Client) DeleteToken(ctx context.Context, id int) (DeleteTokenResponse, error) {
	data, err := c.sendDeleteRequest(ctx, path.Join(tokensPath, strconv.Itoa(id)))
	if err != nil {
		return DeleteTokenResponse{}, err
	}

	var response DeleteTokenResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return DeleteTokenResponse{}, err
	}

	if !response.Success {
		return response, errors.New("the API request did not succeed")
	}
	return response, nil
}

type GetTokenResponse struct {
	Success bool  `json:"success"`
	Data    Token `json:"data"`
}

func (c *Client) GetToken(ctx context.Context, id int) (GetTokenResponse, error) {
	data, err := c.sendGetRequest(ctx, path.Join(tokensPath, strconv.Itoa(id)), nil)
	if err != nil {
		return GetTokenResponse{}, err
	}

	var response GetTokenResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return GetTokenResponse{}, err
	}

	if !response.Success {
		return response, errors.New("the API request did not succeed")
	}
	return response, nil
}
