package ctfdapi

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"strconv"
)

const (
	challengesPath = "/api/v1/challenges"
)

type Challenge struct {
	// Id is the unique id of the challenge. This field needs to be configured as omitempty. Otherwise, a create call
	// will submit the Id to the API endpoint, which will break database constraints.
	Id             int    `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Attribution    string `json:"attribution,omitempty"`
	ConnectionInfo string `json:"connection_info,omitempty"` //nolint:tagliatelle
	NextId         *int   `json:"next_id,omitempty"`         //nolint:tagliatelle
	MaxAttempts    int    `json:"max_attempts,omitempty"`    //nolint:tagliatelle
	Value          int    `json:"value,omitempty"`
	Category       string `json:"category,omitempty"`
	Type           string `json:"type,omitempty"`
	State          string `json:"state,omitempty"`
}

type ListChallengeResponse struct {
	Success bool        `json:"success"`
	Data    []Challenge `json:"data"`
}

func (c *Client) ListChallenges(ctx context.Context) ([]Challenge, error) {
	data, err := c.sendGetRequest(ctx, challengesPath, map[string]string{"view": "admin"})
	if err != nil {
		return nil, err
	}

	var response ListChallengeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

type CreateChallengeResponse struct {
	Success bool      `json:"success"`
	Data    Challenge `json:"data"`
}

func (c *Client) CreateChallenge(ctx context.Context, challenge Challenge) (Challenge, error) {
	// Creating a challenge with a specific ID will sooner or later result in violated database constraints.
	// To prevent that, we reset the challenge id.
	challenge.Id = 0
	if challenge.Type == "" {
		// If no challenge type is set, the API call will fail on the server side with an internal server error.
		// We are setting the default on the client side to avoid unexpected errors.
		challenge.Type = "standard"
	}
	data, err := c.sendPostRequest(ctx, challengesPath, challenge)
	if err != nil {
		return Challenge{}, err
	}

	var response CreateChallengeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return Challenge{}, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

type DeleteChallengeResponse struct {
	Success bool `json:"success"`
}

func (c *Client) DeleteChallenge(ctx context.Context, id int) error {
	data, err := c.sendDeleteRequest(ctx, path.Join(challengesPath, strconv.Itoa(id)))
	if err != nil {
		return err
	}

	var response DeleteChallengeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	if !response.Success {
		return errors.New("the API request did not succeed")
	}
	return nil
}

type UpdateChallengeResponse struct {
	Success bool      `json:"success"`
	Data    Challenge `json:"data"`
}

func (c *Client) UpdateChallenge(ctx context.Context, challenge Challenge) (Challenge, error) {
	data, err := c.sendPatchRequest(ctx, path.Join(challengesPath, strconv.Itoa(challenge.Id)), challenge)
	if err != nil {
		return Challenge{}, err
	}

	var response UpdateChallengeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return Challenge{}, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

type GetChallengeResponse struct {
	Success bool      `json:"success"`
	Data    Challenge `json:"data"`
}

func (c *Client) GetChallenge(ctx context.Context, id int) (Challenge, error) {
	data, err := c.sendGetRequest(ctx, path.Join(challengesPath, strconv.Itoa(id)), nil)
	if err != nil {
		return Challenge{}, err
	}

	var response GetChallengeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return Challenge{}, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}
