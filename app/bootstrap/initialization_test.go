package bootstrap

import (
	"testing"

	"glance-sentry-releases/app/config"
)

func TestInitEnv(t *testing.T) {
	t.Setenv("SENTRY_ORG", "acme")
	t.Setenv("SENTRY_AUTH_TOKEN", "token")

	env := config.NewEnv()
	if err := InitEnv(env); err != nil {
		t.Fatalf("InitEnv() error = %v", err)
	}

	if env.SentryOrg != "acme" || env.SentryToken != "token" {
		t.Errorf("env = %+v, want org acme and token token", env)
	}
}

func TestValidateEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     *config.Env
		wantErr bool
	}{
		{"valid", &config.Env{SentryOrg: "acme", SentryToken: "token"}, false},
		{"incomplete", &config.Env{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnv(tt.env)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnv() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
