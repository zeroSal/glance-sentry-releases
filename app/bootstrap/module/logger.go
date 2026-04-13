package module

import (
	"glance-sentry-releases/app/service/logger"

	"go.uber.org/fx"
)

type AppLogger struct{ logger.LoggerInterface }

var Logger = fx.Module("logger",
	fx.Provide(logger.NewConsoleLogger),
	fx.Provide(func(l *logger.ConsoleLogger) *AppLogger {
		return &AppLogger{l}
	}),
)
