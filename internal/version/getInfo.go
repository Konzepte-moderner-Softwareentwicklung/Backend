package version

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
)

type Info struct {
	Version       string `json:"version"`
	Commit        string `json:"commit"`
	Branch        string `json:"branch"`
	BuildTime     string `json:"build_time"`
	GoVersion     string `json:"go_version"`
	NumGoroutines int    `json:"num_goroutines"`
}

func LoggerWithVersion(versionJSON string, baseLogger zerolog.Logger) zerolog.Logger {
	fmt.Println(versionJSON)
	var info Info
	baseLogger.Info().Str("version info", versionJSON).Msg("")
	if err := json.Unmarshal([]byte(versionJSON), &info); err != nil {
		baseLogger.Error().Err(err).Str("raw_version", versionJSON).Msg("failed to parse version info")
		return baseLogger
	}
	fmt.Printf("Version: %s\nCommit: %s\nBranch: %s\nBuild Time: %s\n", info.Version, info.Commit, info.Branch, info.BuildTime)
	return baseLogger.With().
		Str("version", info.Version).
		Str("commit", info.Commit).
		Str("branch", info.Branch).
		Str("build_time", info.BuildTime).
		Str("go_version", info.GoVersion).
		Int("go_routines", info.NumGoroutines).
		Logger()
}
