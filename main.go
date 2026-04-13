package main

import (
	"log"

	"glance-sentry-releases/app"
	"glance-sentry-releases/cmd"

	"github.com/spf13/cobra"
)

var (
	Version   = app.Version()
	Channel   = "stable"
	BuildDate = ""
)

func main() {
	buildSpecs := app.NewBuildSpecs(Version, Channel, BuildDate)

	root := &cobra.Command{
		Use:   "glance-sentry-releases",
		Short: "Sentry releases proxy with adoption metrics",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				log.Print(err)
			}
		},
	}

	root.AddCommand(
		cmd.NewServeCmd(buildSpecs).Command(),
	)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
