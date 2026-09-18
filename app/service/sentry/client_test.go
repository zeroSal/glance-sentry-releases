package sentry

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type rewriteTransport struct {
	target *url.URL
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server URL error = %v", err)
	}

	client := NewClient("acme", "secret")
	client.httpClient.Transport = rewriteTransport{target: target}
	return client
}

func TestClientGetProjects(t *testing.T) {
	var gotPath, gotAuth string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`[{"id":"42","slug":"web","name":"Web"}]`))
	})

	projects, err := client.GetProjects()
	if err != nil {
		t.Fatalf("GetProjects() error = %v", err)
	}

	if gotPath != "/api/0/organizations/acme/projects/" {
		t.Errorf("path = %q, want %q", gotPath, "/api/0/organizations/acme/projects/")
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer secret")
	}
	if len(projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(projects))
	}
	if projects[0].ID.String() != "42" || projects[0].Slug != "web" || projects[0].Name != "Web" {
		t.Errorf("projects[0] = %+v, want id 42, slug web, name Web", projects[0])
	}
}

func TestClientGetProjectsNumericID(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":42,"slug":"web","name":"Web"}]`))
	})

	projects, err := client.GetProjects()
	if err != nil {
		t.Fatalf("GetProjects() error = %v", err)
	}

	if projects[0].ID.String() != "42" {
		t.Errorf("ID = %q, want %q", projects[0].ID.String(), "42")
	}
}

func TestClientGetReleases(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`[{"version":"1.0.0","shortVersion":"1.0.0","newGroups":3}]`))
	})

	releases, err := client.GetReleases("42")
	if err != nil {
		t.Fatalf("GetReleases() error = %v", err)
	}

	if gotPath != "/api/0/organizations/acme/releases/" {
		t.Errorf("path = %q, want %q", gotPath, "/api/0/organizations/acme/releases/")
	}

	wantQuery := map[string]string{
		"project":            "42",
		"health":             "1",
		"per_page":           "20",
		"summaryStatsPeriod": "24h",
		"healthStatsPeriod":  "24h",
		"flatten":            "1",
	}
	for key, want := range wantQuery {
		if got := gotQuery.Get(key); got != want {
			t.Errorf("query %q = %q, want %q", key, got, want)
		}
	}

	if len(releases) != 1 {
		t.Fatalf("len(releases) = %d, want 1", len(releases))
	}
	if releases[0].Version != "1.0.0" || releases[0].NewGroups != 3 {
		t.Errorf("releases[0] = %+v, want version 1.0.0 and 3 new groups", releases[0])
	}
}

func TestClientErrorStatus(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"detail":"forbidden"}`))
	})

	_, err := client.GetProjects()
	if err == nil {
		t.Fatal("GetProjects() error = nil, want error")
	}

	want := `sentry organizations/acme/projects/ → 403: {"detail":"forbidden"}`
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestClientInvalidJSON(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	})

	if _, err := client.GetProjects(); err == nil {
		t.Fatal("GetProjects() error = nil, want error")
	}
}

func TestClientTransportError(t *testing.T) {
	client := NewClient("acme", "secret")
	server := httptest.NewServer(http.NotFoundHandler())
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server URL error = %v", err)
	}
	server.Close()
	client.httpClient.Transport = rewriteTransport{target: target}

	_, err = client.GetReleases("42")
	if err == nil {
		t.Fatal("GetReleases() error = nil, want error")
	}
	if strings.Contains(err.Error(), "→") {
		t.Errorf("error = %v, want a transport error", err)
	}
}
