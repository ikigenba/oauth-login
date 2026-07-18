// Package oauth implements the provider-neutral OAuth authorization-code flow
// with PKCE.
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const maxErrorBody = 4096

// Client contains the provider-independent OAuth client configuration.
type Client struct {
	AuthURL      string
	TokenURL     string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scope        string
	Rand         io.Reader
	HTTPClient   *http.Client
}

// Session carries the per-login secrets from authorization through exchange.
type Session struct {
	State        string
	CodeVerifier string
}

// Param is a key/value parameter or HTTP header supplied by the caller.
type Param struct {
	Key   string
	Value string
}

// NewSession creates fresh state and PKCE verifier values.
func (c *Client) NewSession() (*Session, error) {
	random := c.Rand
	if random == nil {
		random = rand.Reader
	}

	verifierBytes := make([]byte, 64)
	if _, err := io.ReadFull(random, verifierBytes); err != nil {
		return nil, fmt.Errorf("generate code verifier: %w", err)
	}
	stateBytes := make([]byte, 32)
	if _, err := io.ReadFull(random, stateBytes); err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	return &Session{
		State:        base64.RawURLEncoding.EncodeToString(stateBytes),
		CodeVerifier: base64.RawURLEncoding.EncodeToString(verifierBytes),
	}, nil
}

var authorizeReserved = map[string]struct{}{
	"response_type":         {},
	"client_id":             {},
	"redirect_uri":          {},
	"state":                 {},
	"code_challenge":        {},
	"code_challenge_method": {},
	"scope":                 {},
}

// AuthorizeURL constructs the browser authorization URL for a session.
func (c *Client) AuthorizeURL(session *Session, extra []Param) (string, error) {
	if session == nil {
		return "", fmt.Errorf("session is required")
	}
	if err := validateVerifier(session.CodeVerifier); err != nil {
		return "", err
	}
	for _, param := range extra {
		if _, reserved := authorizeReserved[param.Key]; reserved {
			if param.Key == "redirect_uri" {
				return "", fmt.Errorf("reserved authorize parameter %q; use --callback-host, --port, and --callback-path to configure it", param.Key)
			}
			return "", fmt.Errorf("reserved authorize parameter %q", param.Key)
		}
	}

	authURL, err := url.Parse(c.AuthURL)
	if err != nil {
		return "", fmt.Errorf("parse authorization URL: %w", err)
	}

	digest := sha256.Sum256([]byte(session.CodeVerifier))
	query := url.Values{
		"response_type":         {"code"},
		"client_id":             {c.ClientID},
		"redirect_uri":          {c.RedirectURI},
		"state":                 {session.State},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(digest[:])},
		"code_challenge_method": {"S256"},
	}
	if c.Scope != "" {
		query.Set("scope", c.Scope)
	}

	extraQuery := url.Values{}
	for _, param := range extra {
		extraQuery.Add(param.Key, param.Value)
	}
	authURL.RawQuery = appendEncoded(query.Encode(), extraQuery.Encode())
	return authURL.String(), nil
}

func validateVerifier(verifier string) error {
	if len(verifier) < 43 || len(verifier) > 128 {
		return fmt.Errorf("code verifier length must be between 43 and 128 characters")
	}
	for _, char := range verifier {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || strings.ContainsRune("-._~", char) {
			continue
		}
		return fmt.Errorf("code verifier contains invalid character %q", char)
	}
	return nil
}

var exchangeReserved = map[string]struct{}{
	"grant_type":    {},
	"code":          {},
	"code_verifier": {},
	"redirect_uri":  {},
	"client_id":     {},
	"client_secret": {},
}

// Exchange trades an authorization code for the token endpoint's raw response.
func (c *Client) Exchange(ctx context.Context, session *Session, code string, extra []Param, headers []Param) ([]byte, error) {
	if session == nil {
		return nil, fmt.Errorf("session is required")
	}
	for _, param := range extra {
		if _, reserved := exchangeReserved[param.Key]; reserved {
			return nil, fmt.Errorf("reserved token parameter %q", param.Key)
		}
	}

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {session.CodeVerifier},
		"redirect_uri":  {c.RedirectURI},
		"client_id":     {c.ClientID},
	}
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}
	extraForm := url.Values{}
	for _, param := range extra {
		extraForm.Add(param.Key, param.Value)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL, strings.NewReader(appendEncoded(form.Encode(), extraForm.Encode())))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, header := range headers {
		request.Header.Add(header.Key, header.Value)
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		displayBody := body
		if len(displayBody) > maxErrorBody {
			displayBody = displayBody[:maxErrorBody]
		}
		return nil, fmt.Errorf("token endpoint returned %s: %s", response.Status, displayBody)
	}

	return body, nil
}

func appendEncoded(required, extra string) string {
	if extra == "" {
		return required
	}
	if required == "" {
		return extra
	}
	return required + "&" + extra
}
