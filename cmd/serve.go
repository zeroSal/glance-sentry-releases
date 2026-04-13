package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"glance-sentry-releases/app"
	"glance-sentry-releases/app/bootstrap/module"
	"glance-sentry-releases/app/config"
	"glance-sentry-releases/app/model/release"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type ServeCmd struct {
	buildSpecs *app.BuildSpecs
}

func NewServeCmd(buildSpecs *app.BuildSpecs) *ServeCmd {
	return &ServeCmd{buildSpecs: buildSpecs}
}

func (s *ServeCmd) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the Sentry releases proxy server",
		Run:   s.run,
	}
}

func (s *ServeCmd) run(cmd *cobra.Command, args []string) {
	a := fx.New(
		fx.NopLogger,
		fx.Provide(func() *app.BuildSpecs { return s.buildSpecs }),
		app.Kernel,
		fx.Invoke(s.execute),
	)
	a.Run()
}

func (s *ServeCmd) execute(
	env *config.Env,
	sentryClient *module.SentryClient,
	log *module.AppLogger,
) {
	if err := env.Validate(); err != nil {
		log.Error("Validation failed: " + err.Error())
		return
	}

	addr := env.GetProxyAddr()
	log.Info("Glance Sentry proxy is listening on " + addr)
	http.HandleFunc("/", s.createHandler(sentryClient, log))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Error("Server error: " + err.Error())
	}
}

func (s *ServeCmd) createHandler(sentryClient *module.SentryClient, log *module.AppLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := sentryClient.GetProjects()
		if err != nil {
			http.Error(w, "projects: "+err.Error(), http.StatusInternalServerError)
			log.Error("Failed to fetch projects: " + err.Error())
			return
		}

		result := make([]release.ProjectOut, 0, len(projects))

		for _, proj := range projects {
			releases, err := sentryClient.GetReleases(proj.ID.String())
			if err != nil {
				log.Warn("Failed to fetch releases for " + proj.Slug + ": " + err.Error())
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

				adoption, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", adoption), 64)

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

		resp := release.Response{
			FetchedAt: time.Now().Format("15:04"),
			Projects:  result,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("JSON encode error: " + err.Error())
		}
	}
}
