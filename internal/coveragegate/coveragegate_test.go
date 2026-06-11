package coveragegate

import (
	"go/build/constraint"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// validReasons is the CLOSED taxonomy of no-test allowlist reason codes. A
// bare path or an unknown code fails the gate (fail-closed) — the allowlist
// cannot be silently weakened.
var validReasons = map[string]string{
	"non-windows-platform": "build-tagged out on Windows (unix/!windows/linux/darwin); macOS/Linux is upstream Bumblebee's domain, out of Pollen's scope",
	"type-only":            "pure type/const declarations with no executable statements",
}

// coverageAllowlist maps repo-relative (forward-slash) production .go files to
// a reason code in validReasons. Only files that genuinely cannot or need not
// be tested on Pollen's Windows target belong here.
var coverageAllowlist = map[string]string{
	// Unix inode/device symlink-loop key — compiled only on unix GOOSes.
	"internal/walk/dirkey_unix.go": "non-windows-platform",
	// macOS/Linux root discovery for cmd/pollen — //go:build !windows.
	"cmd/pollen/roots_notwindows.go": "non-windows-platform",
}

// scanDirsSkipped are top-level directories with no testable Go packages.
var scanDirsSkipped = map[string]struct{}{
	"dist": {}, "docs": {}, "threat_intel": {}, ".git": {}, ".github": {},
}

// TestAllowlistIsReasonCoded fails closed: every allowlist entry must use a
// known reason code and point at a file that exists.
func TestAllowlistIsReasonCoded(t *testing.T) {
	root := repoRoot(t)
	for rel, reason := range coverageAllowlist {
		if _, ok := validReasons[reason]; !ok {
			t.Errorf("allowlist %q uses unknown reason code %q — fail-closed (valid: %v)", rel, reason, reasonCodes())
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("allowlist entry %q does not exist: %v", rel, err)
		}
	}
}

// TestProductionFilesLinkedOrAllowlisted walks every production .go file and
// asserts: a Windows-compiled file's package has a test OR the file is
// allowlisted; a non-Windows file MUST be allowlisted (so a new platform file
// cannot slip in untracked).
func TestProductionFilesLinkedOrAllowlisted(t *testing.T) {
	root := repoRoot(t)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if _, skip := scanDirsSkipped[d.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		rel := toSlashRel(t, root, path)

		if _, ok := coverageAllowlist[rel]; ok {
			return nil // validated by TestAllowlistIsReasonCoded
		}

		if !compiledOnWindows(t, path) {
			t.Errorf("%s is build-tagged out on Windows but is NOT in the coverage allowlist — add it with reason \"non-windows-platform\" (fail-closed)", rel)
			return nil
		}

		if !packageHasTest(filepath.Dir(path)) {
			t.Errorf("%s: package has no _test.go and is not allowlisted — Windows-compiled code must be package-tested", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func reasonCodes() []string {
	out := make([]string, 0, len(validReasons))
	for k := range validReasons {
		out = append(out, k)
	}
	return out
}

// compiledOnWindows reports whether the Go toolchain compiles path when
// GOOS=windows, honoring both the platform filename suffix convention and an
// explicit //go:build constraint.
func compiledOnWindows(t *testing.T, path string) bool {
	t.Helper()
	stem := strings.TrimSuffix(filepath.Base(path), ".go")
	for _, suf := range []string{"_unix", "_linux", "_darwin", "_notwindows", "_android", "_ios", "_freebsd", "_openbsd"} {
		if strings.HasSuffix(stem, suf) {
			return false
		}
	}
	if strings.HasSuffix(stem, "_windows") {
		return true
	}
	expr := buildConstraint(t, path)
	if expr == nil {
		return true // unconstrained → compiled everywhere, including Windows
	}
	return expr.Eval(func(tag string) bool { return tag == "windows" })
}

// buildConstraint returns the parsed //go:build expression of path, or nil if
// the file has none. It only scans the leading comment/blank block.
func buildConstraint(t *testing.T, path string) constraint.Expr {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if constraint.IsGoBuild(trimmed) {
			expr, perr := constraint.Parse(trimmed)
			if perr != nil {
				t.Fatalf("parse build constraint in %s: %v", path, perr)
			}
			return expr
		}
		// Build constraints must precede the package clause; stop at the first
		// line that is neither blank nor a comment.
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
			break
		}
	}
	return nil
}

func packageHasTest(dir string) bool {
	matches, _ := filepath.Glob(filepath.Join(dir, "*_test.go"))
	return len(matches) > 0
}

func toSlashRel(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatalf("rel(%s,%s): %v", root, path, err)
	}
	return filepath.ToSlash(rel)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test directory")
		}
		dir = parent
	}
}
