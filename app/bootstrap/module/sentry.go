package module

import (
	"glance-sentry-releases/app/config"
	"glance-sentry-releases/app/service/sentry"

	"go.uber.org/fx"
)

type SentryClient struct{ sentry.ClientInterface }

var Sentry = fx.Module("sentry",
	fx.Provide(func(env *config.Env) *sentry.Client {
		return sentry.NewClient(env.SentryOrg, env.SentryToken)
	}),
	fx.Provide(func(c *sentry.Client) *SentryClient {
		return &SentryClient{c}
	}),
)
