package cmd

import (
	"glance-sentry-releases/app"

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
	)
	a.Run()
}
