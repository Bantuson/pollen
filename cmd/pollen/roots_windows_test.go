//go:build windows

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bantuson/pollen/internal/model"
)

// TestWindowsBaselineRoots proves that resolveRoots(baseline) returns roots for
// all 8 package ecosystems when fake env-var dirs exist under t.TempDir.
// It satisfies WRES-01 (JS ecosystems) and WRES-02 (PyPI/Go/RubyGems/Composer).
func TestWindowsBaselineRoots(t *testing.T) {
	tmp := t.TempDir()
	appdata := filepath.Join(tmp, "AppData", "Roaming")
	localappdata := filepath.Join(tmp, "AppData", "Local")
	programFiles := filepath.Join(tmp, "ProgramFiles")

	t.Setenv("USERPROFILE", tmp)
	t.Setenv("APPDATA", appdata)
	t.Setenv("LOCALAPPDATA", localappdata)
	t.Setenv("ProgramFiles", programFiles)

	// Create the concrete directories each ecosystem root expects.
	// Glob roots (Python*, .gem/ruby/*, Ruby*) need a concrete versioned dir
	// under the wildcard parent so globExisting finds them.
	mustMkdir := func(p string) {
		t.Helper()
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("MkdirAll %q: %v", p, err)
		}
	}

	// npm user global modules
	npmModules := filepath.Join(appdata, "npm", "node_modules")
	mustMkdir(npmModules)

	// pnpm content-addressable store
	pnpmStore := filepath.Join(localappdata, "pnpm", "store")
	mustMkdir(pnpmStore)

	// Yarn global install tree
	yarnGlobal := filepath.Join(localappdata, "Yarn", "Data", "global")
	mustMkdir(yarnGlobal)

	// Bun install cache
	bunCache := filepath.Join(tmp, ".bun", "install", "cache")
	mustMkdir(bunCache)

	// PyPI user site-packages — create a concrete Python313 versioned dir
	pypiRoot := filepath.Join(appdata, "Python", "Python313", "site-packages")
	mustMkdir(pypiRoot)

	// Go modules cache
	goMod := filepath.Join(tmp, "go", "pkg", "mod")
	mustMkdir(goMod)

	// RubyGems user gems — create a concrete 3.3.0 versioned dir
	rubyGemsUser := filepath.Join(tmp, ".gem", "ruby", "3.3.0", "gems")
	mustMkdir(rubyGemsUser)

	// Composer global vendor
	composerVendor := filepath.Join(appdata, "Composer", "vendor")
	mustMkdir(composerVendor)

	// npm global via Node.js MSI (ProgramFiles)
	npmMSI := filepath.Join(programFiles, "nodejs", "node_modules")
	mustMkdir(npmMSI)

	// RubyGems system — create a concrete Ruby33-x64 versioned dir
	rubyGemsSystem := filepath.Join(programFiles, "Ruby33-x64", "lib", "ruby", "gems")
	mustMkdir(rubyGemsSystem)

	roots, _, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{})
	if err != nil {
		t.Fatalf("resolveRoots baseline: %v", err)
	}

	// Build path→kind map for O(1) lookup.
	got := make(map[string]string, len(roots))
	for _, r := range roots {
		got[r.Path] = r.Kind
	}

	// Assert every created dir is present with the expected Kind.
	assertRoot := func(path, wantKind string) {
		t.Helper()
		kind, ok := got[path]
		if !ok {
			t.Errorf("baseline missing root %q (got %v)", path, roots)
			return
		}
		if kind != wantKind {
			t.Errorf("root %q kind = %q, want %q", path, kind, wantKind)
		}
	}

	// npm user global modules — RootKindGlobalPackage (npm global prefix)
	assertRoot(npmModules, model.RootKindGlobalPackage)
	// pnpm
	assertRoot(pnpmStore, model.RootKindUserPackage)
	// Yarn
	assertRoot(yarnGlobal, model.RootKindUserPackage)
	// Bun
	assertRoot(bunCache, model.RootKindUserPackage)
	// PyPI (glob resolved to versioned dir)
	assertRoot(pypiRoot, model.RootKindUserPackage)
	// Go modules
	assertRoot(goMod, model.RootKindUserPackage)
	// RubyGems user (glob resolved to versioned dir)
	assertRoot(rubyGemsUser, model.RootKindUserPackage)
	// Composer
	assertRoot(composerVendor, model.RootKindUserPackage)
	// npm MSI global (ProgramFiles) — RootKindGlobalPackage
	assertRoot(npmMSI, model.RootKindGlobalPackage)
	// RubyGems system (glob resolved to versioned dir) — RootKindGlobalPackage
	assertRoot(rubyGemsSystem, model.RootKindGlobalPackage)
}

// TestWindowsBaselineRootsEmptyAppdata proves that an empty APPDATA env var
// produces zero APPDATA-derived roots and no CWD-relative path leak.
// This validates the T-02-01 / T-02-06 threat mitigation (Pitfall 1 prevention).
func TestWindowsBaselineRootsEmptyAppdata(t *testing.T) {
	tmp := t.TempDir()

	// Provide USERPROFILE with go/pkg/mod so resolveRoots has at least one
	// surviving root and does not fail with "no default roots".
	t.Setenv("USERPROFILE", tmp)
	goMod := filepath.Join(tmp, "go", "pkg", "mod")
	if err := os.MkdirAll(goMod, 0o755); err != nil {
		t.Fatalf("MkdirAll go mod: %v", err)
	}

	// Empty APPDATA — all APPDATA-derived roots must be suppressed.
	t.Setenv("APPDATA", "")
	// Also clear LOCALAPPDATA to avoid any cross-contamination from the host.
	t.Setenv("LOCALAPPDATA", "")
	t.Setenv("ProgramFiles", "")

	roots, _, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{})
	if err != nil {
		t.Fatalf("resolveRoots baseline with empty APPDATA: %v", err)
	}

	// Assert no returned root has a volume-less path.
	// filepath.VolumeName("npm\\node_modules") == "" — that is the relative-path
	// leak we are guarding against.
	for _, r := range roots {
		if filepath.VolumeName(r.Path) == "" {
			t.Errorf("root %q has no volume name — possible relative-path leak when APPDATA is empty", r.Path)
		}
	}
}
