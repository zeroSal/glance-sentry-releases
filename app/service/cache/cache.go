package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"glance-sentry-releases/app/model/release"
	"glance-sentry-releases/app/service/logger"
	"glance-sentry-releases/app/service/sentry"
)

type Cache struct {
	mu       sync.RWMutex
	data     *release.Response
	fetchErr error
	fetchAt  time.Time
	client   sentry.ClientInterface
	interval time.Duration
	stopChan chan struct{}
	log      logger.LoggerInterface
}

func NewCache(client sentry.ClientInterface, intervalMinutes int, log logger.LoggerInterface) *Cache {
	return &Cache{
		client:   client,
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan struct{}),
		log:      log,
	}
}

func (c *Cache) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-c.stopChan:
				return
			default:
				c.fetch(context.Background())
				if c.data != nil {
					return
				}
				select {
				case <-time.After(5 * time.Second):
				case <-c.stopChan:
					return
				}
			}
		}
	}()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log := c.log
			log.Info("Running scheduled cache refresh...")
			c.fetch(context.Background())
		case <-c.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (c *Cache) WaitForData(ctx context.Context) {
	for {
		c.mu.RLock()
		hasData := c.data != nil
		c.mu.RUnlock()
		if hasData {
			return
		}
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			return
		}
	}
}

func (c *Cache) Stop() {
	close(c.stopChan)
}

func (c *Cache) fetch(ctx context.Context) {
	projects, err := c.client.GetProjects()
	if err != nil {
		c.mu.Lock()
		c.fetchErr = err
		if c.data == nil {
			c.data = &release.Response{
				Projects: []release.ProjectOut{},
			}
		}
		c.data.FetchedAt = time.Now().Format("15:04")
		c.mu.Unlock()
		c.log.Error("Failed to fetch projects: " + err.Error())
		c.log.Warn("Will retry in 5s, using stale/empty cache")
		return
	}

	if len(projects) == 0 {
		c.mu.Lock()
		if c.data == nil {
			c.data = &release.Response{
				Projects: []release.ProjectOut{},
			}
		}
		c.data.FetchedAt = time.Now().Format("15:04")
		c.fetchErr = nil
		c.fetchAt = time.Now()
		c.mu.Unlock()
		c.log.Warn("No projects found")
		return
	}

	result := make([]release.ProjectOut, 0, len(projects))

	for _, proj := range projects {
		select {
		case <-ctx.Done():
			return
		default:
		}

		releases, err := c.client.GetReleases(proj.ID.String())
		if err != nil {
			c.log.Warn("Failed to fetch releases for " + proj.Slug + ": " + err.Error())
			continue
		}

		for _, rel := range releases {
			var adoption float64
			if len(rel.Projects) > 0 && rel.Projects[0].HealthData != nil {
				adoption = rel.Projects[0].HealthData.SessionsAdoption
			}
			if adoption <= 0 {
				continue
			}

			ver := rel.ShortVersion
			if rel.VersionInfo != nil && rel.VersionInfo.Description != "" {
				ver = rel.VersionInfo.Description
			}

			result = append(result, release.ProjectOut{
				Project: proj.Slug,
				Name:    proj.Name,
				Release: release.ReleaseOut{
					Version:     ver,
					FullVersion: rel.Version,
					Adoption:    adoption,
					NewGroups:   rel.NewGroups,
					DateCreated: rel.DateCreated,
				},
			})
			break
		}
	}

	if len(result) == 0 {
		c.mu.Lock()
		if c.data == nil {
			c.data = &release.Response{
				Projects: []release.ProjectOut{},
			}
		}
		c.data.FetchedAt = time.Now().Format("15:04")
		c.fetchErr = fmt.Errorf("no projects with releases found")
		c.mu.Unlock()
		c.log.Warn("No projects with valid releases found")
		return
	}

	c.mu.Lock()
	if c.data == nil {
		c.data = &release.Response{}
	}
	c.data.FetchedAt = time.Now().Format("15:04")
	c.data.Projects = result
	c.fetchErr = nil
	c.fetchAt = time.Now()
	c.mu.Unlock()

	c.log.Success(fmt.Sprintf("Cache refreshed: %d projects", len(result)))
}

func (c *Cache) Get() (*release.Response, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.data == nil {
		if c.fetchErr != nil {
			return nil, c.fetchErr
		}
		return nil, nil
	}
	resp := *c.data
	return &resp, nil
}
