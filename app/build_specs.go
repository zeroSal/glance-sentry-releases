package app

import (
	"fmt"
	"runtime/debug"
)

type BuildSpecs struct {
	version   string
	channel   string
	buildDate string
}

func NewBuildSpecs(version, channel, buildDate string) *BuildSpecs {
	return &BuildSpecs{
		version:   version,
		channel:   channel,
		buildDate: buildDate,
	}
}

func (b *BuildSpecs) GetVersion() string {
	if b.version == "" {
		return "dev"
	}
	return b.version
}

func (b *BuildSpecs) GetChannel() string {
	return b.channel
}

func (b *BuildSpecs) GetBuildDate() string {
	return b.buildDate
}

func (b *BuildSpecs) GetFullVersion() string {
	return fmt.Sprintf("%s (%s)", b.GetVersion(), b.channel)
}

func Version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:8]
			}
		}
	}
	return "dev"
}
