package nexus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const DefaultBaseURL = "https://api.nexusmods.com/v1"

// Client is a minimal Nexus V1 REST client.
type Client struct {
	baseURL   string
	apiKey    string
	appName   string
	appVer    string
	httpc     *http.Client
	userAgent string
}

type ClientOption func(*Client)

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithHTTPClient(httpc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpc = httpc
	}
}

func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func NewClient(apiKey, appName, appVersion string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		apiKey:  apiKey,
		appName: appName,
		appVer:  appVersion,
		httpc: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.userAgent == "" {
		c.userAgent = fmt.Sprintf("%s/%s", c.appName, c.appVer)
	}
	return c
}

func (c *Client) doJSON(ctx context.Context, method, fullURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", c.apiKey)
	if c.appName != "" {
		req.Header.Set("Application-Name", c.appName)
	}
	if c.appVer != "" {
		req.Header.Set("Application-Version", c.appVer)
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			URL:        fullURL,
			Body:       string(b),
		}
	}

	if out == nil {
		return nil
	}
	if len(b) == 0 {
		return fmt.Errorf("nexus api: empty response: %s %s", method, fullURL)
	}
	if err := json.Unmarshal(b, out); err != nil {
		const max = 300
		s := string(b)
		if len(s) > max {
			s = s[:max] + "…"
		}
		return fmt.Errorf("nexus api: invalid json for %s %s: %w (body=%q)", method, fullURL, err, s)
	}
	return nil
}

// ValidateUser checks that the apiKey is valid and returns basic user info.
// Endpoint: GET /v1/users/validate.json
func (c *Client) ValidateUser(ctx context.Context) (*ValidateUserResponse, error) {
	u := fmt.Sprintf("%s/users/validate.json", c.baseURL)
	var out ValidateUserResponse
	if err := c.doJSON(ctx, http.MethodGet, u, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SearchMods is intentionally not implemented for REST v1.
func (c *Client) SearchMods(ctx context.Context, gameDomain, q string) ([]SearchResult, error) {
	return nil, fmt.Errorf("nexus search is not supported via Nexus REST v1")
}

// GetMod returns mod details.
// Endpoint: GET /v1/games/{game_domain}/mods/{mod_id}.json
func (c *Client) GetMod(ctx context.Context, gameDomain string, modID int) (*ModInfo, error) {
	u := fmt.Sprintf("%s/games/%s/mods/%d.json", c.baseURL, url.PathEscape(gameDomain), modID)
	var out ModInfo
	if err := c.doJSON(ctx, http.MethodGet, u, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListFiles returns file list for a mod.
// Endpoint: GET /v1/games/{game_domain}/mods/{mod_id}/files.json
func (c *Client) ListFiles(ctx context.Context, gameDomain string, modID int) ([]FileInfo, error) {
	u := fmt.Sprintf("%s/games/%s/mods/%d/files.json", c.baseURL, url.PathEscape(gameDomain), modID)

	var resp struct {
		Files []FileInfo `json:"files"`
	}

	if err := c.doJSON(ctx, http.MethodGet, u, &resp); err != nil {
		return nil, err
	}
	return resp.Files, nil
}

// GetDownloadLinks resolves download links for a specific file.
//
// Endpoint:
//
//	GET /v1/games/{game}/mods/{mod_id}/files/{file_id}/download_link.json?key=...&expires=...&user_id=...
func (c *Client) GetDownloadLinks(ctx context.Context, gameDomain string, modID, fileID int, key, expires, userID string) ([]DownloadLink, error) {
	base := fmt.Sprintf("%s/games/%s/mods/%d/files/%d/download_link.json", c.baseURL, url.PathEscape(gameDomain), modID, fileID)

	q := url.Values{}
	q.Set("key", key)
	q.Set("expires", expires)
	q.Set("user_id", userID)

	u := base + "?" + q.Encode()

	var out []DownloadLink
	if err := c.doJSON(ctx, http.MethodGet, u, &out); err != nil {
		return nil, err
	}
	return out, nil
}
