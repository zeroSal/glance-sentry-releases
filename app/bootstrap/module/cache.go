package module

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"glance-sentry-releases/app/config"
	"glance-sentry-releases/app/service/cache"

	"go.uber.org/fx"
)

type Cache struct{ *cache.Cache }

var CacheModule = fx.Module("cache",
	fx.Provide(func(sc *SentryClient, env *config.Env, log *AppLogger) *cache.Cache {
		return cache.NewCache(sc.ClientInterface, env.CacheIntervalMinutes, log)
	}),
	fx.Invoke(func(lc fx.Lifecycle, c *cache.Cache, env *config.Env, log *AppLogger) {
		log.Info(fmt.Sprintf("Cache interval: %d minutes", env.CacheIntervalMinutes))
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.Info("Starting cache fetcher in background...")
				go c.Start(ctx)
				log.Success("HTTP server ready on " + env.GetProxyAddr())
				go func() {
					http.HandleFunc("/", createHandler(c, log))
					if err := http.ListenAndServe(env.GetProxyAddr(), nil); err != nil {
						log.Error("Server error: " + err.Error())
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				c.Stop()
				return nil
			},
		})
	}),
)

func createHandler(cache *cache.Cache, log *AppLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := cache.Get()
		if err != nil {
			http.Error(w, "cache: "+err.Error(), http.StatusInternalServerError)
			log.Error("Cache error: " + err.Error())
			return
		}
		if data == nil {
			http.Error(w, "cache: no data available yet", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Error("JSON encode error: " + err.Error())
		}
	}
}
