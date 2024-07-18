package main

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	hashFileRe = regexp.MustCompile(`(checksums?|(md5|sha1|sha128|sha256|sha512)(sums?)?)\b`)
	isMusl     = isMuslLibcPresent()
	muslRe     = regexp.MustCompile(`[ -_\.]musl[ -_\.]?\b`)
)

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
		"x64",
		"64bit",
	},
	"arm64": []string{
		"aarch64",
	},
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

func matchesOS(name, goos string) bool {
	return mapMatches(name, goos, knownOSAliases)
}

func matchesArch(name, goarch string) bool {
	return mapMatches(name, goarch, knownArchAliases)
}

func mapMatches(name, target string, targetAliases map[string][]string) bool {
	name = strings.ToLower(name)
	if strings.Contains(name, target) {
		return true
	} else if aliases, ok := targetAliases[target]; ok {
		if containsAlias(name, aliases) {
			return true
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
	return hashFileRe.MatchString(strings.ToLower(name))
}

func containsMusl(name string) bool {
	return muslRe.MatchString(strings.ToLower(name))
}

func isMuslLibcPresent() bool {
	matches, _ := filepath.Glob("/lib/libc.musl-*")
	return len(matches) > 0
}
