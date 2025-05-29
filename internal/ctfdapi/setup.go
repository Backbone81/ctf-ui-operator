package ctfdapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	setupPath = "/setup"
)

type SetupRequest struct {
	// General tab
	CTFName        string
	CTFDescription string

	// Mode tab
	UserMode UserMode

	// Settings tab
	ChallengeVisibility    ChallengeVisibility
	AccountVisibility      AccountVisibility
	ScoreVisibility        ScoreVisibility
	RegistrationVisibility RegistrationVisibility
	VerifyEmails           bool
	TeamSize               *int

	// Administration tab
	Name     string
	Email    string
	Password string

	// Style tab
	CTFTheme   CTFTheme
	ThemeColor *string

	// Date & Time tab
	Start *time.Time
	End   *time.Time
}

type UserMode string

const (
	UserModeTeams UserMode = "teams"
	UserModeUsers UserMode = "users"
)

type ChallengeVisibility string

const (
	ChallengeVisibilityPublic  ChallengeVisibility = "public"
	ChallengeVisibilityPrivate ChallengeVisibility = "private"
	ChallengeVisibilityAdmins  ChallengeVisibility = "admins"
)

type AccountVisibility string

const (
	AccountVisibilityPublic  AccountVisibility = "public"
	AccountVisibilityPrivate AccountVisibility = "private"
	AccountVisibilityAdmins  AccountVisibility = "admins"
)

type ScoreVisibility string

const (
	ScoreVisibilityPublic  ScoreVisibility = "public"
	ScoreVisibilityPrivate ScoreVisibility = "private"
	ScoreVisibilityHidden  ScoreVisibility = "hidden"
	ScoreVisibilityAdmins  ScoreVisibility = "admins"
)

type RegistrationVisibility string

const (
	RegistrationVisibilityPublic  RegistrationVisibility = "public"
	RegistrationVisibilityPrivate RegistrationVisibility = "private"
	RegistrationVisibilityMLC     RegistrationVisibility = "mlc"
)

type CTFTheme string

const (
	CTFThemeCoreBeta CTFTheme = "core-beta"
	CTFThemeCore     CTFTheme = "core"
)

// SetupRequired checks if the setup process was already done or not.
func (c *Client) SetupRequired(ctx context.Context) (bool, error) {
	targetUrl, err := c.getTargetUrl(setupPath)
	if err != nil {
		return false, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetUrl, nil)
	if err != nil {
		return false, fmt.Errorf("creating new HTTP request: %w", err)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return false, fmt.Errorf("executing HTTP request: %w", err)
	}
	defer response.Body.Close() //nolint:errcheck

	_, err = io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Errorf("reading response body: %w", err)
	}

	if response.StatusCode == http.StatusOK {
		return true, nil
	}
	if response.StatusCode == http.StatusFound {
		return false, nil
	}
	return false, fmt.Errorf("unexpected status code %d: %s", response.StatusCode, response.Status)
}

// Setup initializes the CTFd instance. This request is special, as there is no REST API endpoint for setup. Instead,
// we are interacting with the HTML site and send form fields as required.
func (c *Client) Setup(ctx context.Context, setupRequest SetupRequest) error {
	nonce, err := c.setupGetNonce(ctx)
	if err != nil {
		return fmt.Errorf("getting nonce: %w", err)
	}
	if err := c.setupSendForm(ctx, setupRequest, nonce); err != nil {
		return fmt.Errorf("sending form: %w", err)
	}
	return nil
}

var nonceRegex = regexp.MustCompile(`<input id="nonce" name="nonce" type="hidden" value="([^"]+)">`)

// setupGetNonce executes a GET request on the setup endpoint and extracts the nonce from the hidden field of the HTML.
// The nonce is required for sending a POST request. Otherwise, the website will reject the request.
func (c *Client) setupGetNonce(ctx context.Context) (string, error) {
	targetUrl, err := c.getTargetUrl(setupPath)
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

	matches := nonceRegex.FindSubmatch(pageData)
	if len(matches) != 2 {
		return "", errors.New("nonce not found in HTML")
	}
	return string(matches[1]), nil
}

// setupSendForm constructs a POST request to the setup endpoint with the configuration provided by the SetupRequest.
func (c *Client) setupSendForm(ctx context.Context, setupRequest SetupRequest, nonce string) error {
	targetUrl, err := c.getTargetUrl(setupPath)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	multipartWriter := multipart.NewWriter(&body)
	if err := c.setupRequestToForm(multipartWriter, setupRequest); err != nil {
		return err
	}
	if err := multipartWriter.WriteField("_submit", "Finish"); err != nil {
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

//nolint:cyclop,funlen,gocognit // breaking this method up into multiple methods would not make it easier to understand
func (c *Client) setupRequestToForm(writer *multipart.Writer, setupRequest SetupRequest) error {
	// General tab
	if err := writer.WriteField("ctf_name", setupRequest.CTFName); err != nil {
		return err
	}
	if err := writer.WriteField("ctf_description", setupRequest.CTFDescription); err != nil {
		return err
	}

	// Mode tab
	if err := writer.WriteField("user_mode", string(setupRequest.UserMode)); err != nil {
		return err
	}

	// Settings tab
	if err := writer.WriteField("challenge_visibility", string(setupRequest.ChallengeVisibility)); err != nil {
		return err
	}
	if err := writer.WriteField("account_visibility", string(setupRequest.AccountVisibility)); err != nil {
		return err
	}
	if err := writer.WriteField("score_visibility", string(setupRequest.ScoreVisibility)); err != nil {
		return err
	}
	if err := writer.WriteField("registration_visibility", string(setupRequest.RegistrationVisibility)); err != nil {
		return err
	}
	if err := writer.WriteField("verify_emails", strconv.FormatBool(setupRequest.VerifyEmails)); err != nil {
		return err
	}
	if setupRequest.TeamSize != nil {
		if err := writer.WriteField("team_size", strconv.Itoa(*setupRequest.TeamSize)); err != nil {
			return err
		}
	} else {
		if err := writer.WriteField("team_size", ""); err != nil {
			return err
		}
	}

	// Administration tab
	if err := writer.WriteField("name", setupRequest.Name); err != nil {
		return err
	}
	if err := writer.WriteField("email", setupRequest.Email); err != nil {
		return err
	}
	if err := writer.WriteField("password", setupRequest.Password); err != nil {
		return err
	}

	// Style tab
	if err := writer.WriteField("ctf_theme", string(setupRequest.CTFTheme)); err != nil {
		return err
	}
	if setupRequest.ThemeColor != nil {
		if err := writer.WriteField("theme_color", *setupRequest.ThemeColor); err != nil {
			return err
		}
	} else {
		if err := writer.WriteField("theme_color", ""); err != nil {
			return err
		}
	}

	// Date & Time tab
	if setupRequest.Start != nil {
		if err := writer.WriteField("start", strconv.FormatInt(setupRequest.Start.Unix(), 10)); err != nil {
			return err
		}
	} else {
		if err := writer.WriteField("start", ""); err != nil {
			return err
		}
	}
	if setupRequest.End != nil {
		if err := writer.WriteField("end", strconv.FormatInt(setupRequest.End.Unix(), 10)); err != nil {
			return err
		}
	} else {
		if err := writer.WriteField("end", ""); err != nil {
			return err
		}
	}
	return nil
}
