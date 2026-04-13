package sentry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"glance-sentry-releases/app/model/release"
)

var _ ClientInterface = (*Client)(nil)

type Client struct {
	sentryOrg   string
	sentryToken string
}

func NewClient(sentryOrg, sentryToken string) *Client {
	return &Client{
		sentryOrg:   sentryOrg,
		sentryToken: sentryToken,
	}
}

func (c *Client) GetProjects() ([]release.Project, error) {
	var projects []release.Project
	err := c.get("organizations/"+c.sentryOrg+"/projects/", &projects)
	return projects, err
}

func (c *Client) GetReleases(projectID string) ([]release.Release, error) {
	url := fmt.Sprintf(
		"organizations/%s/releases/?project=%s&health=1&per_page=20"+
			"&summaryStatsPeriod=24h&healthStatsPeriod=24h&flatten=1",
		c.sentryOrg, projectID,
	)
	var releases []release.Release
	err := c.get(url, &releases)
	return releases, err
}

func (c *Client) get(path string, out any) error {
	req, err := http.NewRequest("GET", "https://sentry.io/api/0/"+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.sentryToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sentry %s → %d: %s", path, resp.StatusCode, body)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
