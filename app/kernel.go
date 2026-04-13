package app

import (
	"glance-sentry-releases/app/bootstrap"
	"glance-sentry-releases/app/bootstrap/module"
	"go.uber.org/fx"
)

var Kernel = fx.Module(
	"app",
	bootstrap.Module,
	module.Logger,
	module.Sentry,
)
