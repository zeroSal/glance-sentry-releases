package config

import "testing"

func TestEnvLoadDefaults(t *testing.T) {
	t.Setenv("SENTRY_ORG", "acme")
	t.Setenv("SENTRY_AUTH_TOKEN", "token")
	t.Setenv("GLANCE_SENTRY_PORT", "")
	t.Setenv("GLANCE_SENTRY_HOST", "")
	t.Setenv("CACHE_INTERVAL_MINUTES", "")

	env := NewEnv()
	if err := env.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if env.SentryOrg != "acme" {
		t.Errorf("SentryOrg = %q, want %q", env.SentryOrg, "acme")
	}
	if env.SentryToken != "token" {
		t.Errorf("SentryToken = %q, want %q", env.SentryToken, "token")
	}
	if env.GlanceSentryPort != "8099" {
		t.Errorf("GlanceSentryPort = %q, want %q", env.GlanceSentryPort, "8099")
	}
	if env.GlanceSentryHost != "127.0.0.1" {
		t.Errorf("GlanceSentryHost = %q, want %q", env.GlanceSentryHost, "127.0.0.1")
	}
	if env.CacheIntervalMinutes != 5 {
		t.Errorf("CacheIntervalMinutes = %d, want %d", env.CacheIntervalMinutes, 5)
	}
}

func TestEnvLoadOverrides(t *testing.T) {
	t.Setenv("SENTRY_ORG", "acme")
	t.Setenv("SENTRY_AUTH_TOKEN", "token")
	t.Setenv("GLANCE_SENTRY_PORT", "9000")
	t.Setenv("GLANCE_SENTRY_HOST", "0.0.0.0")
	t.Setenv("CACHE_INTERVAL_MINUTES", "15")

	env := NewEnv()
	if err := env.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if env.GlanceSentryPort != "9000" {
		t.Errorf("GlanceSentryPort = %q, want %q", env.GlanceSentryPort, "9000")
	}
	if env.GlanceSentryHost != "0.0.0.0" {
		t.Errorf("GlanceSentryHost = %q, want %q", env.GlanceSentryHost, "0.0.0.0")
	}
	if env.CacheIntervalMinutes != 15 {
		t.Errorf("CacheIntervalMinutes = %d, want %d", env.CacheIntervalMinutes, 15)
	}
}

func TestEnvLoadInvalidCacheInterval(t *testing.T) {
	t.Setenv("CACHE_INTERVAL_MINUTES", "not-a-number")

	env := NewEnv()
	if err := env.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if env.CacheIntervalMinutes != 5 {
		t.Errorf("CacheIntervalMinutes = %d, want %d", env.CacheIntervalMinutes, 5)
	}
}

func TestEnvValidate(t *testing.T) {
	tests := []struct {
		name    string
		env     Env
		wantErr string
	}{
		{
			name: "valid",
			env:  Env{SentryOrg: "acme", SentryToken: "token"},
		},
		{
			name:    "missing org",
			env:     Env{SentryToken: "token"},
			wantErr: "SENTRY_ORG is required",
		},
		{
			name:    "missing token",
			env:     Env{SentryOrg: "acme"},
			wantErr: "SENTRY_AUTH_TOKEN is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("Validate() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestEnvGetProxyAddr(t *testing.T) {
	env := Env{GlanceSentryHost: "127.0.0.1", GlanceSentryPort: "8099"}

	if got := env.GetProxyAddr(); got != "127.0.0.1:8099" {
		t.Errorf("GetProxyAddr() = %q, want %q", got, "127.0.0.1:8099")
	}
}
