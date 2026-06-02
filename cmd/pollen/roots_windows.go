//go:build windows

package main

import (
	"os"
	"path/filepath"

	"github.com/bantuson/pollen/internal/model"
	"github.com/bantuson/pollen/internal/scanner"
)

// windowsBaselinePackageRoots returns the Windows-specific per-user package-manager
// install roots for the baseline profile. All paths are constructed via
// filepath.Join + os.Getenv — never hand-built backslash strings. Absent roots
// are silently dropped by the caller's filterExistingRoots invocation.
//
// Every env-var read is guarded at the variable level so that an empty or unset
// APPDATA / LOCALAPPDATA / USERPROFILE never produces a relative CWD-relative
// path via filepath.Join("", ...) (RESEARCH Q5 Pitfall 1 / T-02-01).
func windowsBaselinePackageRoots() []scanner.Root {
	var out []scanner.Root
	add := func(p, kind string) {
		if p != "" {
			out = append(out, scanner.Root{Path: p, Kind: kind})
		}
	}

	// npm — global user module tree and user-level caches (WRES-01).
	// %APPDATA%\npm\node_modules  — npm global install prefix
	// %APPDATA%\npm-cache\_cacache and %LOCALAPPDATA%\npm-cache\_cacache
	// Both cache locations are included; filterExistingRoots drops whichever
	// is absent on a given machine.
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		add(filepath.Join(appdata, "npm", "node_modules"), model.RootKindGlobalPackage)
		add(filepath.Join(appdata, "npm-cache", "_cacache"), model.RootKindUserPackage)
	}
	if localappdata := os.Getenv("LOCALAPPDATA"); localappdata != "" {
		add(filepath.Join(localappdata, "npm-cache", "_cacache"), model.RootKindUserPackage)
	}

	// pnpm — content-addressable store and global symlink tree (WRES-01).
	if localappdata := os.Getenv("LOCALAPPDATA"); localappdata != "" {
		add(filepath.Join(localappdata, "pnpm", "store"), model.RootKindUserPackage)
	}
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		add(filepath.Join(appdata, "pnpm"), model.RootKindUserPackage)
	}

	// Yarn — global package install tree (WRES-01).
	if localappdata := os.Getenv("LOCALAPPDATA"); localappdata != "" {
		add(filepath.Join(localappdata, "Yarn", "Data", "global"), model.RootKindUserPackage)
	}

	// Bun — install cache (WRES-01).
	if userprofile := os.Getenv("USERPROFILE"); userprofile != "" {
		add(filepath.Join(userprofile, ".bun", "install", "cache"), model.RootKindUserPackage)
	}

	// PyPI — user site-packages (glob: Python*\site-packages) (WRES-02).
	// Wildcard expansion covers multiple CPython minor version installs.
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		for _, p := range globExisting(filepath.Join(appdata, "Python", "Python*", "site-packages")) {
			add(p, model.RootKindUserPackage)
		}
	}

	// Go modules cache (WRES-02).
	// Mirror the Unix default (~/.go/pkg/mod): use %USERPROFILE%\go\pkg\mod.
	// GOPATH override is intentionally not consulted — matches upstream Unix
	// behavior (RESEARCH Open Question 3).
	if userprofile := os.Getenv("USERPROFILE"); userprofile != "" {
		add(filepath.Join(userprofile, "go", "pkg", "mod"), model.RootKindUserPackage)
	}

	// RubyGems — user gem install tree (glob: .gem\ruby\*\gems) (WRES-02).
	if userprofile := os.Getenv("USERPROFILE"); userprofile != "" {
		for _, p := range globExisting(filepath.Join(userprofile, ".gem", "ruby", "*", "gems")) {
			add(p, model.RootKindUserPackage)
		}
	}

	// Composer — global vendor directory (WRES-02).
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		add(filepath.Join(appdata, "Composer", "vendor"), model.RootKindUserPackage)
	}

	return out
}

// windowsSystemRoots returns Windows global/system package roots for the
// baseline profile. These are independent of the current user's home directory
// and are emitted once per scan regardless of --all-users.
func windowsSystemRoots() []scanner.Root {
	var out []scanner.Root
	add := func(p, kind string) {
		if p != "" {
			out = append(out, scanner.Root{Path: p, Kind: kind})
		}
	}

	if programFiles := os.Getenv("ProgramFiles"); programFiles != "" {
		// npm — Node.js MSI installs the global node_modules here (WRES-01).
		add(filepath.Join(programFiles, "nodejs", "node_modules"), model.RootKindGlobalPackage)

		// RubyGems — system Ruby installs (glob: Ruby*\lib\ruby\gems) (WRES-02).
		// Wildcard expansion covers multiple Ruby version installs side-by-side.
		for _, p := range globExisting(filepath.Join(programFiles, "Ruby*", "lib", "ruby", "gems")) {
			add(p, model.RootKindGlobalPackage)
		}
	}

	return out
}
