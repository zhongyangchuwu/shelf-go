package app

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var releaseVersion string

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
