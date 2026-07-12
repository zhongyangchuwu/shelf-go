package app

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// releaseVersion is the local milestone version and may be overridden with -ldflags -X for a distribution build.
var releaseVersion = "v0.2.0"

func String() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev go" + runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH
	}

	ver := releaseVersion
	if ver == "" {
		ver = info.Main.Version
	}
	if ver == "" || ver == "(devel)" {
		ver = vcsRevision(info)
	}

	return fmt.Sprintf("%s %s %s/%s", ver, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func vcsRevision(info *debug.BuildInfo) string {
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if len(s.Value) >= 7 {
				return "dev-" + s.Value[:7]
			}
		}
	}
	return "dev"
}
