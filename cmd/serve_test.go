package cmd

import (
	"testing"

	"glance-sentry-releases/app"
)

func TestServeCmdCommand(t *testing.T) {
	command := NewServeCmd(app.NewBuildSpecs("1.2.3", "stable", "2026-01-01")).Command()

	if command.Use != "serve" {
		t.Errorf("Use = %q, want %q", command.Use, "serve")
	}
	if command.Short != "Start the Sentry releases proxy server" {
		t.Errorf("Short = %q, want %q", command.Short, "Start the Sentry releases proxy server")
	}
	if command.Run == nil {
		t.Error("Run is nil, want the serve handler")
	}
}
