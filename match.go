package main

import (
	"runtime"
	"strings"
)

func matchAsset(name string) bool {
	var matchedOS, matchedArch bool
	lowerName := strings.ToLower(name)
	if strings.Contains(lowerName, runtime.GOOS) {
		matchedOS = true
	}
	if strings.Contains(lowerName, runtime.GOARCH) {
		matchedArch = true
	} else if runtime.GOARCH == "amd64" && strings.Contains(lowerName, "x86_64") {
		matchedArch = true
	}

	return matchedOS && matchedArch
}
