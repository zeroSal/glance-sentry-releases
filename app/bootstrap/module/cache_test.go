package module

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"glance-sentry-releases/app/model/release"
	"glance-sentry-releases/app/service/cache"
)

type fakeSentryClient struct {
	projects []release.Project
	releases []release.Release
}

func (c fakeSentryClient) GetProjects() ([]release.Project, error) {
	return c.projects, nil
}

func (c fakeSentryClient) GetReleases(projectID string) ([]release.Release, error) {
	return c.releases, nil
}

type fakeLogger struct{}

func (fakeLogger) GetIdentifier() string { return "fake" }
func (fakeLogger) Debug(string)          {}
func (fakeLogger) Info(string)           {}
func (fakeLogger) Warn(string)           {}
func (fakeLogger) Error(string)          {}
func (fakeLogger) Success(string)        {}
func (fakeLogger) List([]string)         {}

func TestCreateHandlerServesCachedData(t *testing.T) {
	client := fakeSentryClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: []release.Release{{
			Version:      "abcdef1234",
			ShortVersion: "abcdef1",
			NewGroups:    7,
			DateCreated:  "2026-01-01T00:00:00Z",
			Projects: []release.ReleaseProject{
				{HealthData: &release.HealthData{SessionsAdoption: 12.5}},
			},
		}},
	}
	log := &AppLogger{fakeLogger{}}
	c := cache.NewCache(client, 1, log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go c.Start(ctx)
	defer c.Stop()
	c.WaitForData(ctx)

	recorder := httptest.NewRecorder()
	createHandler(c, log)(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	var resp release.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(resp.Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(resp.Projects))
	}
	if resp.Projects[0].Project != "web" || resp.Projects[0].Release.Adoption != 12.5 {
		t.Errorf("Projects[0] = %+v, want project web with adoption 12.5", resp.Projects[0])
	}
}

func TestCreateHandlerWithoutData(t *testing.T) {
	log := &AppLogger{fakeLogger{}}
	c := cache.NewCache(fakeSentryClient{}, 1, log)

	recorder := httptest.NewRecorder()
	createHandler(c, log)(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if got := recorder.Body.String(); got != "cache: no data available yet\n" {
		t.Errorf("body = %q, want %q", got, "cache: no data available yet\n")
	}
}
