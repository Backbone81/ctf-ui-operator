package ctfdapi

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"strconv"
)

const (
	hintsPath = "/api/v1/hints"
)

//nolint:tagliatelle // This is an externally controlled data type.
type Hint struct {
	// Id is the unique id of the hint. This field needs to be configured as omitempty. Otherwise, a create call
	// will submit the Id to the API endpoint, which will break database constraints.
	Id          int    `json:"id,omitempty"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	ChallengeId int    `json:"challenge_id"`
	Content     string `json:"content"`
	Cost        int    `json:"cost"`
}

type ListHintsResponse struct {
	Success bool   `json:"success"`
	Data    []Hint `json:"data"`
}

func (c *Client) ListHintsForChallenge(ctx context.Context, challengeId int) ([]Hint, error) {
	data, err := c.sendGetRequest(ctx, hintsPath, map[string]string{"challenge_id": strconv.Itoa(challengeId)})
	if err != nil {
		return nil, err
	}

	var response ListHintsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

func (c *Client) ListHints(ctx context.Context) ([]Hint, error) {
	data, err := c.sendGetRequest(ctx, hintsPath, nil)
	if err != nil {
		return nil, err
	}

	var response ListHintsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

type CreateHintResponse struct {
	Success bool `json:"success"`
	Data    Hint `json:"data"`
}

func (c *Client) CreateHint(ctx context.Context, hint Hint) (Hint, error) {
	// Creating a hint with a specific ID will sooner or later result in violated database constraints.
	// To prevent that, we reset the hint id.
	hint.Id = 0
	if hint.Type == "" {
		// If no hint type is set, the API call will fail on the server side with an internal server error.
		// We are setting the default on the client side to avoid unexpected errors.
		hint.Type = "standard"
	}
	data, err := c.sendPostRequest(ctx, hintsPath, hint)
	if err != nil {
		return Hint{}, err
	}

	var response CreateHintResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return Hint{}, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

type DeleteHintResponse struct {
	Success bool `json:"success"`
}

func (c *Client) DeleteHint(ctx context.Context, id int) error {
	data, err := c.sendDeleteRequest(ctx, path.Join(hintsPath, strconv.Itoa(id)))
	if err != nil {
		return err
	}

	var response DeleteHintResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	if !response.Success {
		return errors.New("the API request did not succeed")
	}
	return nil
}

type UpdateHintResponse struct {
	Success bool `json:"success"`
	Data    Hint `json:"data"`
}

func (c *Client) UpdateHint(ctx context.Context, hint Hint) (Hint, error) {
	data, err := c.sendPatchRequest(ctx, path.Join(hintsPath, strconv.Itoa(hint.Id)), hint)
	if err != nil {
		return Hint{}, err
	}

	var response UpdateHintResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return Hint{}, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

type GetHintResponse struct {
	Success bool `json:"success"`
	Data    Hint `json:"data"`
}

func (c *Client) GetHint(ctx context.Context, id int) (Hint, error) {
	data, err := c.sendGetRequest(ctx, path.Join(hintsPath, strconv.Itoa(id)), nil)
	if err != nil {
		return Hint{}, err
	}

	var response GetHintResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return Hint{}, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}
