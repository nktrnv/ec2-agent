package imds

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const baseURL = "http://169.254.169.254/latest/meta-data"

type Metadata struct {
	Username string
	Password string
	SSHKey   string
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
	}
}

func (c *Client) Fetch() (Metadata, error) {
	username, err := c.fetch("username")
	if err != nil {
		return Metadata{}, fmt.Errorf("username: %w", err)
	}

	password, err := c.fetch("password")
	if err != nil {
		return Metadata{}, fmt.Errorf("password: %w", err)
	}

	sshKey, err := c.fetch("public-keys/0/openssh-key")
	if err != nil {
		return Metadata{}, fmt.Errorf("ssh key: %w", err)
	}

	return Metadata{
		Username: username,
		Password: password,
		SSHKey:   sshKey,
	}, nil
}

func (c *Client) fetch(path string) (result string, err error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/"+path, nil)
	if err != nil {
		return "", fmt.Errorf("create request for %q: %w", path, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %q: %w", path, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close response for %q: %w", path, closeErr)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response for %q: %w", path, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %q: unexpected status %s", path, resp.Status)
	}

	value := strings.TrimSpace(string(body))
	if value == "" {
		return "", nil
	}

	return value, nil
}
