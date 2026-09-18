package release

import (
	"encoding/json"
	"testing"
)

func TestProjectUnmarshal(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantID string
	}{
		{"string id", `{"id":"42","slug":"web","name":"Web"}`, "42"},
		{"numeric id", `{"id":42,"slug":"web","name":"Web"}`, "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var project Project
			if err := json.Unmarshal([]byte(tt.input), &project); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if project.ID.String() != tt.wantID {
				t.Errorf("ID = %q, want %q", project.ID.String(), tt.wantID)
			}
			if project.Slug != "web" || project.Name != "Web" {
				t.Errorf("project = %+v, want slug web and name Web", project)
			}
		})
	}
}

func TestReleaseUnmarshal(t *testing.T) {
	input := `{
		"version": "abcdef1234",
		"shortVersion": "abcdef1",
		"versionInfo": {"description": "1.2.3"},
		"newGroups": 7,
		"dateCreated": "2026-01-01T00:00:00Z",
		"projects": [{"healthData": {"sessionsAdoption": 12.5}}]
	}`

	var rel Release
	if err := json.Unmarshal([]byte(input), &rel); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if rel.Version != "abcdef1234" || rel.ShortVersion != "abcdef1" {
		t.Errorf("versions = %q / %q, want abcdef1234 / abcdef1", rel.Version, rel.ShortVersion)
	}
	if rel.VersionInfo == nil || rel.VersionInfo.Description != "1.2.3" {
		t.Errorf("VersionInfo = %+v, want description 1.2.3", rel.VersionInfo)
	}
	if rel.NewGroups != 7 {
		t.Errorf("NewGroups = %d, want 7", rel.NewGroups)
	}
	if len(rel.Projects) != 1 || rel.Projects[0].HealthData == nil {
		t.Fatalf("Projects = %+v, want one entry with health data", rel.Projects)
	}
	if got := rel.Projects[0].HealthData.SessionsAdoption; got != 12.5 {
		t.Errorf("SessionsAdoption = %v, want 12.5", got)
	}
}

func TestReleaseUnmarshalWithoutHealthData(t *testing.T) {
	var rel Release
	if err := json.Unmarshal([]byte(`{"version":"1.0.0","projects":[{}]}`), &rel); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if rel.VersionInfo != nil {
		t.Errorf("VersionInfo = %+v, want nil", rel.VersionInfo)
	}
	if len(rel.Projects) != 1 || rel.Projects[0].HealthData != nil {
		t.Errorf("Projects = %+v, want one entry without health data", rel.Projects)
	}
}

func TestResponseMarshal(t *testing.T) {
	resp := Response{
		FetchedAt: "10:30",
		Projects: []ProjectOut{
			{
				Project: "web",
				Name:    "Web",
				Release: ReleaseOut{
					Version:     "1.2.3",
					FullVersion: "abcdef1234",
					Adoption:    12.5,
					NewGroups:   7,
					DateCreated: "2026-01-01T00:00:00Z",
				},
			},
		},
	}

	out, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	want := `{"fetchedAt":"10:30","projects":[{"project":"web","name":"Web",` +
		`"release":{"version":"1.2.3","fullVersion":"abcdef1234","adoption":12.5,` +
		`"newGroups":7,"dateCreated":"2026-01-01T00:00:00Z"}}]}`
	if string(out) != want {
		t.Errorf("Marshal() = %s, want %s", out, want)
	}
}
