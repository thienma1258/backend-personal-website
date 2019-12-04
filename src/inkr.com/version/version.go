package version

import (
	"fmt"
	"runtime"
)

// Version - The version of the app.
// This is inferred from the Git branch (release or hotfix) in pipeline/debug.
var Version = "0.0.1"

// BuildHash - The Git commit hash.
var BuildHash = "08323e0"

// BuildDate - The time when build happens.
var BuildDate = "2019-06-24T10:44:00Z"

// BuildPlatform - The runtime platform.
var BuildPlatform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

// GoVersion - The runtime Go version.
var GoVersion = runtime.Version()
