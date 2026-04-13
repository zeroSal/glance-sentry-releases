package sentry

import (
	"glance-sentry-releases/app/model/release"
)

type ClientInterface interface {
	GetProjects() ([]release.Project, error)
	GetReleases(projectID string) ([]release.Release, error)
}
