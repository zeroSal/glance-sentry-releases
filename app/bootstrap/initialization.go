package bootstrap

import (
	"glance-sentry-releases/app/config"
	"go.uber.org/fx"
)

func InitEnv(env *config.Env) error {
	return env.Load()
}

func ValidateEnv(env *config.Env) error {
	return env.Validate()
}

var Module = fx.Module(
	"bootstrap",
	fx.Provide(config.NewEnv),
	fx.Invoke(InitEnv),
	fx.Invoke(ValidateEnv),
)
