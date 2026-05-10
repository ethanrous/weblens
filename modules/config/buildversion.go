package config

import (
	"os"
	"runtime/debug"
)

const unknownBuildVersion = "unknown"

// BuildVersion returns the build version of the running binary. It checks the
// WEBLENS_BUILD_VERSION environment variable first, falling back to the
// embedded VCS revision from runtime/debug. Returns "unknown" if neither is
// available.
func BuildVersion() string {
	if v := os.Getenv("WEBLENS_BUILD_VERSION"); v != "" {
		return v
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return unknownBuildVersion
	}

	for _, s := range buildInfo.Settings {
		if s.Key == "vcs.revision" {
			return s.Value
		}
	}

	return unknownBuildVersion
}
