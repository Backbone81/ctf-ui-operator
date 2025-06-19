package ctfdapi

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"strconv"
)

const (
	flagsPath = "/api/v1/flags"
)

//nolint:tagliatelle // This is an externally controlled data type.
type Flag struct {
	// Id is the unique id of the flag. This field needs to be configured as omitempty. Otherwise, a create call
	// will submit the Id to the API endpoint, which will break database constraints.
	Id          int    `json:"id,omitempty"`
	ChallengeId int    `json:"challenge_id"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Data        string `json:"data"`
}

type ListFlagsResponse struct {
	Success bool   `json:"success"`
	Data    []Flag `json:"data"`
}

func (c *Client) ListFlags(ctx context.Context) ([]Flag, error) {
	data, err := c.sendGetRequest(ctx, flagsPath, nil)
	if err != nil {
		return nil, err
	}

	var response ListFlagsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

func (c *Client) ListFlagsForChallenge(ctx context.Context, challengeId int) ([]Flag, error) {
	data, err := c.sendGetRequest(ctx, flagsPath, map[string]string{"challenge_id": strconv.Itoa(challengeId)})
	if err != nil {
		return nil, err
	}

	var response ListFlagsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

type CreateFlagResponse struct {
	Success bool `json:"success"`
	Data    Flag `json:"data"`
}

func (c *Client) CreateFlag(ctx context.Context, flag Flag) (Flag, error) {
	// Creating a flag with a specific ID will sooner or later result in violated database constraints.
	// To prevent that, we reset the flag id.
	flag.Id = 0
	if flag.Type == "" {
		// If no flag type is set, the API call will fail on the server side with an internal server error.
		// We are setting the default on the client side to avoid unexpected errors.
		flag.Type = "static"
	}
	data, err := c.sendPostRequest(ctx, flagsPath, flag)
	if err != nil {
		return Flag{}, err
	}

	var response CreateFlagResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return Flag{}, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

type DeleteFlagResponse struct {
	Success bool `json:"success"`
}

func (c *Client) DeleteFlag(ctx context.Context, id int) error {
	data, err := c.sendDeleteRequest(ctx, path.Join(flagsPath, strconv.Itoa(id)))
	if err != nil {
		return err
	}

	var response DeleteFlagResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	if !response.Success {
		return errors.New("the API request did not succeed")
	}
	return nil
}

type UpdateFlagResponse struct {
	Success bool `json:"success"`
	Data    Flag `json:"data"`
}

func (c *Client) UpdateFlag(ctx context.Context, flag Flag) (Flag, error) {
	data, err := c.sendPatchRequest(ctx, path.Join(flagsPath, strconv.Itoa(flag.Id)), flag)
	if err != nil {
		return Flag{}, err
	}

	var response UpdateFlagResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return Flag{}, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}

type GetFlagResponse struct {
	Success bool `json:"success"`
	Data    Flag `json:"data"`
}

func (c *Client) GetFlag(ctx context.Context, id int) (Flag, error) {
	data, err := c.sendGetRequest(ctx, path.Join(flagsPath, strconv.Itoa(id)), nil)
	if err != nil {
		return Flag{}, err
	}

	var response GetFlagResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return Flag{}, err
	}

	if !response.Success {
		return response.Data, errors.New("the API request did not succeed")
	}
	return response.Data, nil
}
