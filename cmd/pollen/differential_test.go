package main

// differential_test.go — TestDifferential: the load-bearing PTEST-02 guard.
//
// NAME LOCK: This function is EXACTLY named `TestDifferential`. Plan 04's CI
// invokes it via `go test ./cmd/pollen/ -run '^TestDifferential$'` with zero
// name-drift risk. Do NOT rename, suffix, or parameterize this function.
//
// PURPOSE: Assert that pollen and upstream bumblebee produce byte-for-byte
// identical NDJSON on the fixed diff-fixture AFTER normalization. This proves
// DETECTION-LOGIC parity — same packages found, same ecosystem, same version.
// It does NOT assert self-identification-string parity (scanner_name differs
// by design; see normalize_diff.go for the documented carve-out).
//
// SKIP SEMANTICS:
//   - Windows: TestDifferential always skips with a structured Phase-2 reason.
//     Windows scanner behavior (root resolver, path representation) arrives in
//     Phase 2 (v0.1.1-pollen.2). Diff is Linux+macOS only by design.
//   - Offline developer box (CI != "true"): a clone/build failure is t.Skip
//     with a clear network reason. Acceptable to skip locally without network.
//   - CI runner (CI == "true"): a clone/build failure is t.Fatalf — a TEST
//     FAILURE. If upstream is moved, deleted, or unreachable, the differential
//     guard must fail loudly in CI rather than silently vanishing behind a
//     "normal offline skip". An undetected upstream move would hollow out this
//     guard entirely.
//
// This asymmetry is intentional and documented at the clone site below.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// TestDifferential is the LOCKED-NAME load-bearing differential guard (PTEST-02).
//
// It builds the upstream bumblebee binary from the pinned SHA, runs both pollen
// and upstream bumblebee against a fixed fixture, normalizes both NDJSON outputs
// (stripping the 7 non-deterministic fields + scanner_name), and asserts
// byte-for-byte equality.
//
// Name is LOCKED to `TestDifferential` — plan 04's CI invokes this by exact regexp.
// Do NOT rename, suffix, or split this function.
func TestDifferential(t *testing.T) {
	// Structured Windows skip: differential is Linux+macOS only.
	// Windows root-resolver and path-representation behavior arrives in Phase 2
	// (v0.1.1-pollen.2). The differential guard for Windows will be added then.
	if runtime.GOOS == "windows" {
		t.Skip("PTEST-02 differential runs on Linux+macOS only; Windows behavior arrives Phase 2 (v0.1.1-pollen.2)")
	}

	// Locate the fixed fixture directory (committed, machine-independent).
	// Use the source file path so the test works regardless of cwd.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed — cannot locate testdata")
	}
	fixtureDir := filepath.Join(filepath.Dir(thisFile), "testdata", "diff-fixture")
	if _, err := os.Stat(fixtureDir); err != nil {
		t.Fatalf("testdata/diff-fixture not found at %q: %v", fixtureDir, err)
	}

	// Build upstream bumblebee from the pinned commit SHA.
	//
	// SKIP-vs-FAIL asymmetry (security requirement T-01-10):
	//   - CI == "true"  → clone/build failure is t.Fatalf (TEST FAILURE, not a skip).
	//     A networked CI runner always has network. If upstream is moved, deleted,
	//     or unreachable in CI the guard must FAIL loudly — not silently vanish
	//     behind a "normal offline skip" that masks real drift.
	//   - CI != "true"  → clone/build failure is t.Skip (acceptable: developer
	//     machine may be offline; the guard runs in every CI job where it matters).
	//
	// This comment documents the asymmetry intentionally so reviewers can audit it.
	bumblebeeExe := buildUpstreamBumblebee(t)

	// Build the current pollen binary from this working tree.
	pollenExe := buildCurrentPollen(t)

	// Run both binaries against the fixed fixture with an explicit --root so
	// the default root resolver (profile-based, machine-specific) never runs.
	// Profile=baseline with --root overrides avoids any machine-local paths.
	pollenNDJSON := runBinaryOnFixture(t, pollenExe, fixtureDir)
	bumblebeeNDJSON := runBinaryOnFixture(t, bumblebeeExe, fixtureDir)

	// Normalize both outputs: strip the 7 non-deterministic fields + scanner_name
	// (documented fork divergence), then sort by record_id.
	// normalize() is defined in normalize_diff.go and tested in normalize_diff_test.go.
	pollenNorm, err := normalize(pollenNDJSON)
	if err != nil {
		t.Fatalf("normalize(pollen output): %v\nraw output:\n%s", err, pollenNDJSON)
	}
	bumblebeeNorm, err := normalize(bumblebeeNDJSON)
	if err != nil {
		t.Fatalf("normalize(bumblebee output): %v\nraw output:\n%s", err, bumblebeeNDJSON)
	}

	// Byte-for-byte comparison of the normalized streams.
	// On mismatch, print a structured diff so failures are debuggable.
	if !bytes.Equal(pollenNorm, bumblebeeNorm) {
		// Print a unified line-level diff for easy debugging.
		pollenLines := strings.Split(strings.TrimSpace(string(pollenNorm)), "\n")
		bumblebeeLines := strings.Split(strings.TrimSpace(string(bumblebeeNorm)), "\n")

		var sb strings.Builder
		sb.WriteString("\n--- pollen (normalized)\n+++ bumblebee (normalized)\n")
		// Build a simple line-diff: show each side with its record_id for context.
		maxLen := len(pollenLines)
		if len(bumblebeeLines) > maxLen {
			maxLen = len(bumblebeeLines)
		}
		for i := 0; i < maxLen; i++ {
			p := ""
			b := ""
			if i < len(pollenLines) {
				p = pollenLines[i]
			}
			if i < len(bumblebeeLines) {
				b = bumblebeeLines[i]
			}
			if p != b {
				sb.WriteString(fmt.Sprintf("line %d\n  pollen:    %s\n  bumblebee: %s\n", i+1, recordSummary(p), recordSummary(b)))
			}
		}
		t.Fatalf("PTEST-02 FAILED: pollen and upstream bumblebee NDJSON differ after normalization"+
			"\n(pollen: %d records, bumblebee: %d records)%s",
			len(pollenLines), len(bumblebeeLines), sb.String())
	}

	// Verify both sides found at least some records (not empty/broken output).
	pollenLines := strings.Split(strings.TrimSpace(string(pollenNorm)), "\n")
	if len(pollenNorm) == 0 || (len(pollenLines) == 1 && pollenLines[0] == "") {
		t.Fatal("PTEST-02: pollen produced empty normalized output — fixture may not have been scanned")
	}

	t.Logf("PTEST-02 PASSED: pollen == upstream bumblebee after normalization (%d records)", len(pollenLines))
}

// TestSelftestThreeFindings asserts that `pollen selftest` emits exactly
// 3 findings (npm, pypi, mcp) and exits 0. This is the PTEST-03 regression
// guard. It runs on ALL operating systems — selftest uses an explicit temp-dir
// root with embedded fixtures, NOT the system root resolver, so Windows is not
// excluded (Pitfall 6 from 01-RESEARCH.md).
func TestSelftestThreeFindings(t *testing.T) {
	// runSelftest is defined in selftest.go. We call it in-process (no binary
	// build needed) so this test runs even before a pollen binary is on PATH.
	code := runSelftest([]string{"--quiet"})
	if code != 0 {
		t.Fatalf("pollen selftest exit code = %d, want 0 (PTEST-03 regression)", code)
	}
	// The 3-finding count is asserted inside runSelftest itself (exits 1 if not 3).
	// Calling it here with the exit-code check is the complete regression guard.
}

// buildUpstreamBumblebee clones upstream bumblebee at the pinned SHA and
// builds the binary. Returns the path to the built executable.
//
// Pinned SHA: c24089804ee66ece4bec6f14638cb98985389cdb (v0.1.1 tag)
//
// SKIP-vs-FAIL: see the comment in TestDifferential for the rationale.
// When CI == "true": clone/build failure is t.Fatalf.
// When CI != "true": clone/build failure is t.Skip with a network reason.
func buildUpstreamBumblebee(t *testing.T) string {
	t.Helper()

	// The pinned upstream commit SHA. This is a HARD constraint: the
	// differential test proves pollen matches this EXACT upstream commit.
	// Never replace with a branch name, tag, or latest — that would turn the
	// guard into a moving-target that can silently pass even if pollen has drifted.
	const pinnedSHA = "c24089804ee66ece4bec6f14638cb98985389cdb" // v0.1.1

	inCI := os.Getenv("CI") == "true"

	cloneDir := t.TempDir()

	// Step 1: Clone upstream. On a developer's offline box, a network failure
	// is expected and acceptable (t.Skip). On a CI runner, it is a hard failure
	// (t.Fatalf) — upstream must be reachable from any GitHub-hosted runner.
	//
	// SECURITY NOTE (T-01-09): we always check out the explicit pinned SHA
	// after cloning, not the default branch. This prevents a compromised or
	// raced upstream HEAD from becoming the comparison oracle.
	cloneOut, err := exec.Command("git", "clone", "--depth=1",
		"https://github.com/perplexityai/bumblebee.git", cloneDir).CombinedOutput()
	if err != nil {
		// skip-vs-fail asymmetry (documented in TestDifferential header):
		if inCI {
			// CI runners always have network. Clone failure here means upstream was
			// moved, deleted, or is unreachable. This must be a loud CI failure —
			// not a silent skip that makes the differential guard invisible.
			t.Fatalf("buildUpstreamBumblebee: git clone failed in CI (upstream unreachable?): %v\n%s", err, cloneOut)
		}
		// On an offline developer machine, skip gracefully.
		t.Skipf("buildUpstreamBumblebee: git clone unavailable offline (CI=false); differential runs in CI: %v", err)
	}

	// Step 2: Fetch and check out the exact pinned SHA.
	// --depth=1 above clones only the tip; we need the full history to checkout
	// an older commit. Do a full fetch if the pinned SHA is not already present.
	if out, err := exec.Command("git", "-C", cloneDir, "fetch", "--unshallow").CombinedOutput(); err != nil {
		// If already fully cloned, --unshallow is a no-op error; ignore ENOENT on shallow.
		// Only fail if it looks like a real network error.
		if !strings.Contains(string(out), "already a complete repository") &&
			!strings.Contains(string(out), "does not apply to a complete repository") {
			if inCI {
				t.Fatalf("buildUpstreamBumblebee: git fetch --unshallow failed in CI: %v\n%s", err, out)
			}
			t.Skipf("buildUpstreamBumblebee: git fetch --unshallow offline: %v", err)
		}
	}

	checkoutOut, err := exec.Command("git", "-C", cloneDir, "checkout", pinnedSHA).CombinedOutput()
	if err != nil {
		if inCI {
			t.Fatalf("buildUpstreamBumblebee: git checkout %s failed in CI: %v\n%s", pinnedSHA, err, checkoutOut)
		}
		t.Skipf("buildUpstreamBumblebee: git checkout %s failed (offline?): %v", pinnedSHA, err)
	}

	// Step 3: Build the upstream bumblebee binary.
	binaryName := "bumblebee-upstream"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(cloneDir, binaryName)

	// Build with explicit working directory set to the cloned repo.
	cmd := exec.Command("go", "build",
		"-o", binaryPath,
		"./cmd/bumblebee",
	)
	cmd.Dir = cloneDir
	buildOut, err := cmd.CombinedOutput()
	if err != nil {
		if inCI {
			t.Fatalf("buildUpstreamBumblebee: go build upstream failed in CI: %v\n%s", err, buildOut)
		}
		t.Skipf("buildUpstreamBumblebee: go build upstream failed (missing Go toolchain?): %v", err)
	}

	return binaryPath
}

// buildCurrentPollen builds the current pollen binary from the source tree.
// This uses a temp directory to avoid polluting the source tree with build
// artifacts. Returns the path to the built executable.
func buildCurrentPollen(t *testing.T) string {
	t.Helper()

	// Determine the source root by navigating up from this test file.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed — cannot locate pollen source root")
	}
	// thisFile is cmd/pollen/differential_test.go — go up two levels.
	sourceRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	buildDir := t.TempDir()
	binaryName := "pollen-under-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(buildDir, binaryName)

	cmd := exec.Command("go", "build",
		"-o", binaryPath,
		"./cmd/pollen",
	)
	cmd.Dir = sourceRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("buildCurrentPollen: go build ./cmd/pollen failed: %v\n%s", err, out)
	}

	return binaryPath
}

// runBinaryOnFixture runs the given binary (pollen or bumblebee) against the
// fixed fixture directory with explicit --root to bypass machine-local root
// discovery. Returns the captured stdout NDJSON.
//
// Flags used:
//   - --root <fixtureDir>   — explicit root; no profile-based discovery
//   - --profile deep        — deep + explicit root is the canonical "scan this dir" invocation
//   - --emit-summary=false  — suppress scan_summary records (they include end_time and
//     duration_ms which are non-deterministic; normalize() handles them anyway,
//     but omitting them keeps the fixture comparison simpler)
//
// Note: upstream bumblebee uses the same flags as pollen (the fork preserved
// the CLI surface intentionally for compatibility).
func runBinaryOnFixture(t *testing.T, binaryPath, fixtureDir string) []byte {
	t.Helper()

	cmd := exec.Command(binaryPath,
		"scan",
		"--profile", "deep",
		"--root", fixtureDir,
		"--emit-summary=false",
	)
	out, err := cmd.Output()
	if err != nil {
		// Exit code 1 from the scanner (partial scan, some files unreadable) is
		// acceptable as long as some NDJSON was produced. Exit code 2 is a flag
		// error and is always fatal.
		exitErr, isExit := err.(*exec.ExitError)
		if !isExit || exitErr.ExitCode() == 2 {
			t.Fatalf("runBinaryOnFixture(%s): exit %v\nstdout: %s\nstderr: %s",
				binaryPath, err, out, exitErr.Stderr)
		}
		// Exit 1: partial scan. Proceed with whatever output was produced.
		t.Logf("runBinaryOnFixture(%s): exit 1 (partial scan, proceeding): %v", binaryPath, err)
	}

	return out
}

// recordSummary extracts the record_id from a JSON line for display in diff output.
// Returns a shortened form of the line if record_id cannot be parsed.
func recordSummary(line string) string {
	if line == "" {
		return "<empty>"
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		if len(line) > 120 {
			return line[:120] + "..."
		}
		return line
	}
	parts := []string{}
	for _, key := range []string{"record_id", "record_type", "ecosystem", "package_name", "version"} {
		if v, ok := rec[key].(string); ok && v != "" {
			parts = append(parts, key+"="+v)
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}
