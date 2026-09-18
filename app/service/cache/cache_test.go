package cache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"glance-sentry-releases/app/model/release"
)

type fakeClient struct {
	mu           sync.Mutex
	projects     []release.Project
	projectsErr  error
	releases     map[string][]release.Release
	releasesErr  map[string]error
	projectCalls int
	releaseCalls []string
}

func (c *fakeClient) GetProjects() ([]release.Project, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.projectCalls++
	return c.projects, c.projectsErr
}

func (c *fakeClient) GetReleases(projectID string) ([]release.Release, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.releaseCalls = append(c.releaseCalls, projectID)
	if err, ok := c.releasesErr[projectID]; ok {
		return nil, err
	}
	return c.releases[projectID], nil
}

func (c *fakeClient) calls() (int, []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.projectCalls, append([]string(nil), c.releaseCalls...)
}

type fakeLogger struct{}

func (fakeLogger) GetIdentifier() string { return "fake" }
func (fakeLogger) Debug(string)          {}
func (fakeLogger) Info(string)           {}
func (fakeLogger) Warn(string)           {}
func (fakeLogger) Error(string)          {}
func (fakeLogger) Success(string)        {}
func (fakeLogger) List([]string)         {}

func newCache(client *fakeClient) *Cache {
	return NewCache(client, 1, fakeLogger{})
}

func releaseWithAdoption(version string, adoption float64) release.Release {
	return release.Release{
		Version:      version,
		ShortVersion: version,
		NewGroups:    2,
		DateCreated:  "2026-01-01T00:00:00Z",
		Projects: []release.ReleaseProject{
			{HealthData: &release.HealthData{SessionsAdoption: adoption}},
		},
	}
}

func TestCacheGetWithoutFetch(t *testing.T) {
	data, err := newCache(&fakeClient{}).Get()

	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if data != nil {
		t.Errorf("Get() data = %+v, want nil", data)
	}
}

func TestCacheFetchBuildsResponse(t *testing.T) {
	client := &fakeClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: map[string][]release.Release{
			"1": {releaseWithAdoption("1.0.0", 42.5)},
		},
	}
	c := newCache(client)

	c.fetch(context.Background())

	data, err := c.Get()
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if len(data.Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(data.Projects))
	}
	if data.FetchedAt == "" {
		t.Error("FetchedAt is empty")
	}

	got := data.Projects[0]
	want := release.ProjectOut{
		Project: "web",
		Name:    "Web",
		Release: release.ReleaseOut{
			Version:     "1.0.0",
			FullVersion: "1.0.0",
			Adoption:    42.5,
			NewGroups:   2,
			DateCreated: "2026-01-01T00:00:00Z",
		},
	}
	if got != want {
		t.Errorf("Projects[0] = %+v, want %+v", got, want)
	}
}

func TestCacheFetchPrefersVersionInfoDescription(t *testing.T) {
	rel := releaseWithAdoption("abcdef1234", 10)
	rel.VersionInfo = &release.VersionInfo{Description: "2.0.0"}

	client := &fakeClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: map[string][]release.Release{"1": {rel}},
	}
	c := newCache(client)

	c.fetch(context.Background())

	data, _ := c.Get()
	if got := data.Projects[0].Release.Version; got != "2.0.0" {
		t.Errorf("Version = %q, want %q", got, "2.0.0")
	}
	if got := data.Projects[0].Release.FullVersion; got != "abcdef1234" {
		t.Errorf("FullVersion = %q, want %q", got, "abcdef1234")
	}
}

func TestCacheFetchSkipsReleasesWithoutAdoption(t *testing.T) {
	noHealth := release.Release{Version: "3.0.0", ShortVersion: "3.0.0"}
	client := &fakeClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: map[string][]release.Release{
			"1": {noHealth, releaseWithAdoption("2.0.0", 0), releaseWithAdoption("1.0.0", 5)},
		},
	}
	c := newCache(client)

	c.fetch(context.Background())

	data, _ := c.Get()
	if len(data.Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(data.Projects))
	}
	if got := data.Projects[0].Release.Version; got != "1.0.0" {
		t.Errorf("Version = %q, want %q", got, "1.0.0")
	}
}

func TestCacheFetchKeepsOnlyFirstAdoptedRelease(t *testing.T) {
	client := &fakeClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: map[string][]release.Release{
			"1": {releaseWithAdoption("2.0.0", 80), releaseWithAdoption("1.0.0", 20)},
		},
	}
	c := newCache(client)

	c.fetch(context.Background())

	data, _ := c.Get()
	if len(data.Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(data.Projects))
	}
	if got := data.Projects[0].Release.Version; got != "2.0.0" {
		t.Errorf("Version = %q, want %q", got, "2.0.0")
	}
}

func TestCacheFetchSkipsProjectsWithFailingReleases(t *testing.T) {
	client := &fakeClient{
		projects: []release.Project{
			{ID: "1", Slug: "web", Name: "Web"},
			{ID: "2", Slug: "api", Name: "API"},
		},
		releasesErr: map[string]error{"1": errors.New("boom")},
		releases: map[string][]release.Release{
			"2": {releaseWithAdoption("1.0.0", 30)},
		},
	}
	c := newCache(client)

	c.fetch(context.Background())

	data, _ := c.Get()
	if len(data.Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(data.Projects))
	}
	if got := data.Projects[0].Project; got != "api" {
		t.Errorf("Project = %q, want %q", got, "api")
	}
}

func TestCacheFetchProjectsError(t *testing.T) {
	client := &fakeClient{projectsErr: errors.New("boom")}
	c := newCache(client)

	c.fetch(context.Background())

	data, err := c.Get()
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if data == nil {
		t.Fatal("Get() data = nil, want empty response")
	}
	if len(data.Projects) != 0 {
		t.Errorf("len(Projects) = %d, want 0", len(data.Projects))
	}
	if data.FetchedAt == "" {
		t.Error("FetchedAt is empty")
	}
}

func TestCacheFetchKeepsStaleDataOnError(t *testing.T) {
	client := &fakeClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: map[string][]release.Release{"1": {releaseWithAdoption("1.0.0", 10)}},
	}
	c := newCache(client)
	c.fetch(context.Background())

	client.projectsErr = errors.New("boom")
	c.fetch(context.Background())

	data, err := c.Get()
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if len(data.Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(data.Projects))
	}
	if got := data.Projects[0].Release.Version; got != "1.0.0" {
		t.Errorf("Version = %q, want %q", got, "1.0.0")
	}
}

func TestCacheFetchNoProjects(t *testing.T) {
	c := newCache(&fakeClient{})

	c.fetch(context.Background())

	data, err := c.Get()
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if data == nil || len(data.Projects) != 0 {
		t.Errorf("Get() data = %+v, want empty response", data)
	}
}

func TestCacheFetchNoProjectsWithReleases(t *testing.T) {
	client := &fakeClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: map[string][]release.Release{"1": {releaseWithAdoption("1.0.0", 0)}},
	}
	c := newCache(client)

	c.fetch(context.Background())

	data, err := c.Get()
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if data == nil || len(data.Projects) != 0 {
		t.Errorf("Get() data = %+v, want empty response", data)
	}
}

func TestCacheFetchStopsOnCanceledContext(t *testing.T) {
	client := &fakeClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: map[string][]release.Release{"1": {releaseWithAdoption("1.0.0", 10)}},
	}
	c := newCache(client)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c.fetch(ctx)

	data, err := c.Get()
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if data != nil {
		t.Errorf("Get() data = %+v, want nil", data)
	}

	if _, releaseCalls := client.calls(); len(releaseCalls) != 0 {
		t.Errorf("GetReleases calls = %v, want none", releaseCalls)
	}
}

func TestCacheStartFetchesAndStops(t *testing.T) {
	client := &fakeClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: map[string][]release.Release{"1": {releaseWithAdoption("1.0.0", 10)}},
	}
	c := newCache(client)

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		c.Start(context.Background())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.WaitForData(ctx)

	data, err := c.Get()
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if data == nil || len(data.Projects) != 1 {
		t.Fatalf("Get() data = %+v, want one project", data)
	}

	c.Stop()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("Start() did not return after Stop()")
	}
}

func TestCacheStartStopsOnContextCancel(t *testing.T) {
	c := newCache(&fakeClient{
		projects: []release.Project{{ID: "1", Slug: "web", Name: "Web"}},
		releases: map[string][]release.Release{"1": {releaseWithAdoption("1.0.0", 10)}},
	})
	ctx, cancel := context.WithCancel(context.Background())

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		c.Start(ctx)
	}()

	cancel()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("Start() did not return after context cancel")
	}
}

func TestCacheWaitForDataReturnsOnContextCancel(t *testing.T) {
	c := newCache(&fakeClient{})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		c.WaitForData(ctx)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("WaitForData() did not return after context cancel")
	}
}
