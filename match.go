package main

import (
	"regexp"
	"runtime"
	"strings"
)

var hashFileRe = regexp.MustCompile(`\.(md5|sha1|sha128|sha256|sha512)(sums?)?\b`)

// https://github.com/golang/go/blob/go1.22.5/src/go/build/syslist.go
var knownOS = map[string]bool{
	"aix":       true,
	"android":   true,
	"darwin":    true,
	"dragonfly": true,
	"freebsd":   true,
	"hurd":      true,
	"illumos":   true,
	"ios":       true,
	"js":        true,
	"linux":     true,
	"nacl":      true,
	"netbsd":    true,
	"openbsd":   true,
	"plan9":     true,
	"solaris":   true,
	"wasip1":    true,
	"windows":   true,
	"zos":       true,
}

var knownArch = map[string]bool{
	"386":         true,
	"amd64":       true,
	"amd64p32":    true,
	"arm":         true,
	"armbe":       true,
	"arm64":       true,
	"arm64be":     true,
	"loong64":     true,
	"mips":        true,
	"mipsle":      true,
	"mips64":      true,
	"mips64le":    true,
	"mips64p32":   true,
	"mips64p32le": true,
	"ppc":         true,
	"ppc64":       true,
	"ppc64le":     true,
	"riscv":       true,
	"riscv64":     true,
	"s390":        true,
	"s390x":       true,
	"sparc":       true,
	"sparc64":     true,
	"wasm":        true,
}

var knownOSAliases = map[string][]string{
	"darwin": []string{
		"mac",
	},
	"linux": []string{
		"lnx",
	},
	"windows": []string{
		"win",
	},
}

var knownArchAliases = map[string][]string{
	"amd64": []string{
		"x86_64",
		"64bit",
	},
	"arm64": []string{
		"aarch64",
	},
}

func MatchAsset(name string) bool {
	return matchAsset(name, runtime.GOOS, runtime.GOARCH)
}

func matchAsset(name, goos, goarch string) bool {
	var matchedOS, matchedArch bool
	lowerName := strings.ToLower(name)

	if isIgnoredFile(name) {
		return false
	}

	containedOS := containsOS(name)
	containedArch := containsArch(name)

	if containedOS {
		if strings.Contains(lowerName, goos) {
			matchedOS = true
		} else if aliases, ok := knownOSAliases[goos]; ok {
			if containsAlias(lowerName, aliases) {
				matchedOS = true
			}
		}
	}

	if containedArch {
		if strings.Contains(lowerName, goarch) {
			matchedArch = true
		} else if aliases, ok := knownArchAliases[goarch]; ok {
			if containsAlias(lowerName, aliases) {
				matchedArch = true
			}
		}
	}

	if matchedOS && matchedArch {
		// node_exporter-1.8.2.linux-arm64.tar.gz
		return true
	}

	if !containedOS {
		if containedArch && matchedArch {
			// aerospike-prometheus-exporter_1.17.0_x86_64.tgz
			return true
		}
	} else {
		if matchedOS && !containedArch {
			// prometheus-exporter-linux-1.0.1.tgz
			return true
		}
	}

	return false
}

func containsOS(name string) bool {
	return mapContains(name, knownOS, knownOSAliases)
}

func containsArch(name string) bool {
	return mapContains(name, knownArch, knownArchAliases)
}

func mapContains(name string, m map[string]bool, aliases map[string][]string) bool {
	name = strings.ToLower(name)
	for k, _ := range m {
		if strings.Contains(name, k) {
			return true
		}
	}

	for _, v := range aliases {
		for _, alias := range v {
			if strings.Contains(name, alias) {
				return true
			}
		}
	}

	return false
}

func containsAlias(name string, aliases []string) bool {
	name = strings.ToLower(name)
	for _, v := range aliases {
		if strings.Contains(name, v) {
			return true
		}
	}

	return false
}

func isIgnoredFile(name string) bool {
	name = strings.ToLower(name)
	if hashFileRe.MatchString(name) {
		return true
	}

	return false
}
