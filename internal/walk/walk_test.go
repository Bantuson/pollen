package walk

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestDefaultExcludesCoverProtectedMacOSLibraryPaths ensures the macOS
// Library subtrees that routinely produce TCC denials under broad
// $HOME scans are matched by the default suffix-component excludes.
// Adding new paths to DefaultExcludes is cheap; regressing one of
// these silently is what makes the diagnostics output scary.
func TestDefaultExcludesCoverProtectedMacOSLibraryPaths(t *testing.T) {
	want := []string{
		"Library/ContainerManager",
		"Library/Daemon Containers",
		"Library/DoNotDisturb",
		"Library/DuetExpertCenter",
		"Library/IntelligencePlatform",
		"Library/Photos",
		"Library/Sharing",
		"Library/Shortcuts",
		"Library/StatusKit",
	}
	have := make(map[string]bool, len(DefaultExcludes))
	for _, x := range DefaultExcludes {
		have[x] = true
	}
	for _, w := range want {
		if !have[w] {
			t.Errorf("DefaultExcludes missing %q", w)
		}
	}
}

// TestWalkSkipsExcludedLibrarySubtrees verifies that an exclude with
// a "/"-separated suffix (e.g. "Library/ContainerManager") prunes a
// matching directory anywhere under any root, while a sibling
// directory that does not match continues to be walked.
func TestWalkSkipsExcludedLibrarySubtrees(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("path-separator semantics differ on Windows")
	}
	root := t.TempDir()
	// Simulate a $HOME-shaped tree.
	mustMkdir(t, filepath.Join(root, "Library", "ContainerManager", "deep"))
	mustMkdir(t, filepath.Join(root, "Library", "StatusKit"))
	mustMkdir(t, filepath.Join(root, "code", "proj"))

	// Drop sentinel files we can detect from the visitor.
	mustWrite(t, filepath.Join(root, "Library", "ContainerManager", "deep", "secret.json"), "{}")
	mustWrite(t, filepath.Join(root, "Library", "StatusKit", "x"), "{}")
	mustWrite(t, filepath.Join(root, "code", "proj", "package-lock.json"), "{}")

	excludes := append([]string{}, DefaultExcludes...)

	var seen []string
	err := Walk(Options{
		Roots:    []string{root},
		Excludes: excludes,
	}, func(path string, d fs.DirEntry) error {
		if !d.IsDir() {
			seen = append(seen, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	for _, p := range seen {
		if filepath.Base(filepath.Dir(p)) == "deep" || filepath.Base(filepath.Dir(p)) == "StatusKit" {
			t.Errorf("excluded path was visited: %s", p)
		}
	}
	want := filepath.Join(root, "code", "proj", "package-lock.json")
	found := false
	for _, p := range seen {
		if p == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to visit %q; saw %v", want, seen)
	}
}

// TestWalkExcludesAndVisitorSkip exercises the core walker on ALL platforms
// (unlike TestWalkSkipsExcludedLibrarySubtrees, which skips on Windows). It
// builds the tree with filepath.Join so the separators are OS-native, and
// covers: basename excludes, suffix-component excludes, the empty/whitespace
// branch of normalizeExcludes, dirKey, and a Visitor-returned ErrSkip pruning
// a subtree.
func TestWalkExcludesAndVisitorSkip(t *testing.T) {
	root := t.TempDir()

	// mustWrite does not create parent dirs (the macOS test pairs it with
	// mustMkdir), so create each file's parent before writing.
	write := func(rel ...string) {
		t.Helper()
		p := filepath.Join(append([]string{root}, rel...)...)
		mustMkdir(t, filepath.Dir(p))
		mustWrite(t, p, "{}")
	}
	write("keep", "pkg.json")
	write(".git", "config")              // basename exclude
	write("nested", "cache", "data.txt") // suffix-component exclude
	write("nested", "keep2", "ok.json")
	write("skipme", "inner", "deep.json") // pruned by visitor ErrSkip

	// Forward-slash excludes + an empty and whitespace entry; normalizeExcludes
	// trims the blanks and filepath.Clean rewrites separators to OS-native.
	excludes := []string{".git", "nested/cache", "", "   "}

	var seenFiles []string
	err := Walk(Options{Roots: []string{root}, Excludes: excludes},
		func(path string, d fs.DirEntry) error {
			if d.IsDir() {
				if d.Name() == "skipme" {
					return ErrSkip // prune this subtree via the visitor path
				}
				return nil
			}
			seenFiles = append(seenFiles, path)
			return nil
		})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	visited := func(rel ...string) bool {
		want := filepath.Join(append([]string{root}, rel...)...)
		for _, p := range seenFiles {
			if p == want {
				return true
			}
		}
		return false
	}

	if !visited("keep", "pkg.json") {
		t.Errorf("expected to visit keep/pkg.json; saw %v", seenFiles)
	}
	if !visited("nested", "keep2", "ok.json") {
		t.Errorf("expected to visit nested/keep2/ok.json; saw %v", seenFiles)
	}
	for _, p := range seenFiles {
		base := filepath.Base(p)
		if base == "config" {
			t.Errorf(".git subtree was not excluded: %s", p)
		}
		if base == "data.txt" {
			t.Errorf("nested/cache suffix-component exclude failed: %s", p)
		}
		if base == "deep.json" {
			t.Errorf("skipme subtree was not pruned by visitor ErrSkip: %s", p)
		}
	}
}

// TestWalkOnErrorNonexistentRoot covers the error branch of walkOne's WalkDir
// callback: a missing root surfaces through OnError and the walk still
// completes without error.
func TestWalkOnErrorNonexistentRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	var errs int
	err := Walk(Options{
		Roots:   []string{missing},
		OnError: func(string, error) { errs++ },
	}, func(string, fs.DirEntry) error { return nil })
	if err != nil {
		t.Fatalf("Walk returned error for missing root (should be non-fatal): %v", err)
	}
	if errs == 0 {
		t.Error("OnError was not invoked for a non-existent root")
	}
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, body string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
