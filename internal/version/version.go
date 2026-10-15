// Package version holds build-time identity for the dudes CLI.
package version

// Version is the semantic version of the engine binary.
// Override at link time: -ldflags "-X github.com/Thomas-Pietsch/dudes/internal/version.Version=…"
var Version = "0.0.0-dev"
