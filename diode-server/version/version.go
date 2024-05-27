package version

import (
	_ "embed"
	"strings"
)

// Version is the version of the diode-server
//
//go:embed BUILD_VERSION.txt
var buildVersion string

// Commit is the commit of the diode-server
//
//go:embed BUILD_COMMIT.txt
var buildCommit string

// GetBuildVersion returns the build version of the diode-server
func GetBuildVersion() string {
	return strings.TrimSpace(buildVersion)
}

// GetBuildCommit returns the build commit of the diode-server
func GetBuildCommit() string {
	return strings.TrimSpace(buildCommit)
}
