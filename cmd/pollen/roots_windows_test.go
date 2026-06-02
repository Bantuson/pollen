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

// TestWindowsExtensionMCPRoots proves that resolveRoots(baseline) returns roots for:
//   - all five editor-extension directories (WEXT-01: VS Code, Insiders, Cursor, Windsurf, VSCodium .vscode-oss)
//   - browser-extension roots for Chromium-family (Chrome, Brave) and Firefox (WEXT-02)
//   - all five MCP host-config directories (WEXT-03: Claude, Cline, Cursor, Windsurf, Gemini)
//
// Env vars are set via t.Setenv — never HOME — for Windows test isolation (Phase-2 Pitfall 5 prevention).
// No t.Skip is used; the //go:build windows tag restricts this test to Windows only.
func TestWindowsExtensionMCPRoots(t *testing.T) {
	tmp := t.TempDir()
	appdata := filepath.Join(tmp, "AppData", "Roaming")
	localappdata := filepath.Join(tmp, "AppData", "Local")

	t.Setenv("USERPROFILE", tmp)
	t.Setenv("APPDATA", appdata)
	t.Setenv("LOCALAPPDATA", localappdata)
	t.Setenv("ProgramFiles", filepath.Join(tmp, "ProgramFiles"))

	mustMkdir := func(p string) {
		t.Helper()
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("MkdirAll %q: %v", p, err)
		}
	}

	// WEXT-01: editor-extension roots under USERPROFILE (home).
	vsCode := filepath.Join(tmp, ".vscode", "extensions")
	vsCodeInsiders := filepath.Join(tmp, ".vscode-insiders", "extensions")
	cursor := filepath.Join(tmp, ".cursor", "extensions")
	windsurf := filepath.Join(tmp, ".windsurf", "extensions")
	vscodium := filepath.Join(tmp, ".vscode-oss", "extensions")
	mustMkdir(vsCode)
	mustMkdir(vsCodeInsiders)
	mustMkdir(cursor)
	mustMkdir(windsurf)
	mustMkdir(vscodium)

	// WEXT-02: browser-extension roots under LOCALAPPDATA (Chromium) and APPDATA (Firefox).
	chromeExt := filepath.Join(localappdata, "Google", "Chrome", "User Data", "Default", "Extensions")
	braveExt := filepath.Join(localappdata, "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions")
	firefoxProfiles := filepath.Join(appdata, "Mozilla", "Firefox", "Profiles")
	mustMkdir(chromeExt)
	mustMkdir(braveExt)
	mustMkdir(firefoxProfiles)

	// WEXT-03: MCP config roots — Claude Desktop + Cline under APPDATA; Cursor/Windsurf/Gemini under USERPROFILE.
	claudeDir := filepath.Join(appdata, "Claude")
	clineDir := filepath.Join(appdata, "cline")
	cursorMCP := filepath.Join(tmp, ".cursor")  // already created above as cursor extensions parent
	windsurfMCP := filepath.Join(tmp, ".windsurf") // already created above as windsurf extensions parent
	geminiDir := filepath.Join(tmp, ".gemini")
	// claudeDir and clineDir are new; cursorMCP, windsurfMCP already created.
	mustMkdir(claudeDir)
	mustMkdir(clineDir)
	mustMkdir(geminiDir)

	roots, _, err := resolveRoots(model.ProfileBaseline, nil, rootsOpts{})
	if err != nil {
		t.Fatalf("resolveRoots: %v", err)
	}

	// Build path→kind lookup for assertions.
	got := make(map[string]string, len(roots))
	for _, r := range roots {
		got[r.Path] = r.Kind
	}

	assertRoot := func(path, wantKind string) {
		t.Helper()
		kind, ok := got[path]
		if !ok {
			t.Errorf("baseline missing root %q\n  present roots: %v", path, roots)
			return
		}
		if kind != wantKind {
			t.Errorf("root %q kind = %q, want %q", path, kind, wantKind)
		}
	}

	// WEXT-01: all five editor-extension roots must be present.
	assertRoot(vsCode, model.RootKindEditorExtension)
	assertRoot(vsCodeInsiders, model.RootKindEditorExtension)
	assertRoot(cursor, model.RootKindEditorExtension)
	assertRoot(windsurf, model.RootKindEditorExtension)
	assertRoot(vscodium, model.RootKindEditorExtension)

	// WEXT-02: at least one Chromium-family Extensions dir and the Firefox Profiles parent.
	assertRoot(chromeExt, model.RootKindBrowserExtension)
	assertRoot(braveExt, model.RootKindBrowserExtension)
	assertRoot(firefoxProfiles, model.RootKindBrowserExtension)

	// WEXT-03: all five MCP host-config roots must be present.
	assertRoot(claudeDir, model.RootKindMCPConfig)
	assertRoot(clineDir, model.RootKindMCPConfig)
	assertRoot(cursorMCP, model.RootKindMCPConfig)
	assertRoot(windsurfMCP, model.RootKindMCPConfig)
	assertRoot(geminiDir, model.RootKindMCPConfig)
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
