package release

import (
	"encoding/json"
)

type Project struct {
	ID   json.Number `json:"id"`
	Slug string      `json:"slug"`
	Name string      `json:"name"`
}

type HealthData struct {
	SessionsAdoption float64 `json:"sessionsAdoption"`
}

type ReleaseProject struct {
	HealthData *HealthData `json:"healthData"`
}

type VersionInfo struct {
	Description string `json:"description"`
}

type Release struct {
	Version      string           `json:"version"`
	ShortVersion string           `json:"shortVersion"`
	VersionInfo  *VersionInfo     `json:"versionInfo"`
	NewGroups    int              `json:"newGroups"`
	DateCreated  string           `json:"dateCreated"`
	Projects     []ReleaseProject `json:"projects"`
}

type ReleaseOut struct {
	Version     string  `json:"version"`
	FullVersion string  `json:"fullVersion"`
	Adoption    float64 `json:"adoption"`
	NewGroups   int     `json:"newGroups"`
	DateCreated string  `json:"dateCreated"`
}

type ProjectOut struct {
	Project string     `json:"project"`
	Name    string     `json:"name"`
	Release ReleaseOut `json:"release"`
}

type Response struct {
	FetchedAt string       `json:"fetchedAt"`
	Projects  []ProjectOut `json:"projects"`
}
