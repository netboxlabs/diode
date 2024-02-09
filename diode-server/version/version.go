package version

import _ "embed"

// Version is the version of the diode-server
//
//go:embed VERSION.txt
var Version string
