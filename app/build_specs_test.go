package app

import "testing"

func TestBuildSpecsGetVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"explicit version", "1.2.3", "1.2.3"},
		{"empty falls back to dev", "", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBuildSpecs(tt.version, "stable", "2026-01-01").GetVersion(); got != tt.want {
				t.Errorf("GetVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSpecsAccessors(t *testing.T) {
	specs := NewBuildSpecs("1.2.3", "beta", "2026-01-01")

	if got := specs.GetChannel(); got != "beta" {
		t.Errorf("GetChannel() = %q, want %q", got, "beta")
	}
	if got := specs.GetBuildDate(); got != "2026-01-01" {
		t.Errorf("GetBuildDate() = %q, want %q", got, "2026-01-01")
	}
}

func TestBuildSpecsGetFullVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		channel string
		want    string
	}{
		{"explicit version", "1.2.3", "stable", "1.2.3 (stable)"},
		{"empty version", "", "dev", "dev (dev)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBuildSpecs(tt.version, tt.channel, "").GetFullVersion(); got != tt.want {
				t.Errorf("GetFullVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	if got := Version(); got != "dev" && len(got) != 8 {
		t.Errorf("Version() = %q, want %q or an 8 character revision", got, "dev")
	}
}
