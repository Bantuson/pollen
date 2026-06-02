package main

// parity_test.go — TestParityAllEcosystems: cross-platform parity guard (PTEST-01).
//
// PURPOSE: Prove that pollen's detectors produce equivalent inventory records
// on Linux, macOS, and Windows when scanning the same committed fixture tree.
// The test asserts:
//   1. Every record's endpoint.os equals runtime.GOOS for the running platform.
//   2. At least one record exists for each of the 5 distinct ecosystem strings:
//      npm, pypi, go, rubygems, packagist (covering all 8 package managers in
//      the fixture).
//
// DESIGN: The test drives `pollen scan --profile deep --root testdata/parity-fixture`
// (same mechanism as TestDifferential), so no profile-based root discovery runs —
// only the committed fixture tree is scanned. This makes the test machine-
// independent and identical across all three CI runners.
//
// REUSE: buildCurrentPollen and runBinaryOnFixture are defined in differential_test.go
// (same package main). normalize is defined in normalize_diff.go (same package).
// NONE of those are redefined here. parity-specific helpers are local to this file.
//
// LOCK: normalize() in normalize_diff.go is LOCKED for PTEST-02. This file does
// NOT edit normalize(). endpoint.os handling is local to this file only.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestParityAllEcosystems (PTEST-01): asserts that pollen produces records for
// all 5 ecosystem strings and that every record's endpoint.os matches the
// running OS. No build tag — runs on Linux, macOS, and Windows.
func TestParityAllEcosystems(t *testing.T) {
	// Locate the parity-fixture directory relative to this source file so the
	// test is cwd-independent across all three OS runners.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed — cannot locate testdata")
	}
	fixtureDir := filepath.Join(filepath.Dir(thisFile), "testdata", "parity-fixture")
	if _, err := os.Stat(fixtureDir); err != nil {
		t.Fatalf("testdata/parity-fixture not found at %q: %v", fixtureDir, err)
	}

	// Build the current pollen binary from this working tree.
	// buildCurrentPollen is defined in differential_test.go (same package — do NOT redefine).
	pollenExe := buildCurrentPollen(t)

	// Run pollen against the fixture with explicit --root (no profile-based discovery).
	// runBinaryOnFixture is defined in differential_test.go (same package — do NOT redefine).
	out := runBinaryOnFixture(t, pollenExe, fixtureDir)

	if len(strings.TrimSpace(string(out))) == 0 {
		t.Fatal("PTEST-01: pollen produced empty output — fixture may not have been scanned")
	}

	// Step 1: assert endpoint.os == runtime.GOOS on every record (raw output,
	// before normalization). normalize() does not strip endpoint.os, but asserting
	// on raw is cleanest and avoids any coupling with the normalization path.
	assertEndpointOS(t, out, runtime.GOOS)

	// Step 1b (Windows only): assert Windows path shape and empty uid.
	// Uses RAW out (pre-normalization) so assertions are independent of normalize()'s strip logic.
	// On non-Windows this block is skipped — the Windows helpers are not called.
	if runtime.GOOS == "windows" {
		assertWindowsPathShape(t, out)
		assertWindowsEndpointUID(t, out)
	}

	// Step 2: normalize the output (strip non-deterministic fields, sort by record_id).
	// normalize() is LOCKED — call it unchanged. It keeps endpoint.os in the output.
	norm, err := normalize(out)
	if err != nil {
		t.Fatalf("normalize(pollen output): %v\nraw output:\n%s", err, out)
	}
	if len(strings.TrimSpace(string(norm))) == 0 {
		t.Fatal("PTEST-01: normalized output is empty — fixture may not have been scanned")
	}

	// Step 3: assert ecosystem coverage — at least one record for each of the
	// 5 distinct ecosystem strings that the fixture's 8 package managers emit.
	assertParityRecordCoverage(t, norm)

	normLines := strings.Split(strings.TrimSpace(string(norm)), "\n")
	t.Logf("PTEST-01 PASSED on %s: endpoint.os correct, all 5 ecosystems covered (%d records after normalization)",
		runtime.GOOS, len(normLines))
}

// assertEndpointOS parses each non-blank NDJSON line and fails if any record
// with an endpoint sub-object has endpoint.os != want.
// Called on RAW pollen output (before normalization) so the assertion is
// independent of normalize()'s strip logic.
func assertEndpointOS(t *testing.T, ndjson []byte, want string) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(string(ndjson)), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("assertEndpointOS: line %d: not valid JSON: %v", i+1, err)
			continue
		}
		ep, ok := rec["endpoint"].(map[string]any)
		if !ok {
			// No endpoint sub-object — skip (summary records may not have one).
			continue
		}
		got, _ := ep["os"].(string)
		if got != want {
			t.Errorf("assertEndpointOS: line %d: endpoint.os = %q, want %q", i+1, got, want)
		}
	}
}

// assertWindowsPathShape asserts that every non-empty project_path / source_file
// field in the NDJSON stream contains a drive letter (e.g. "C:") and no forward-slash
// separators. Only called on Windows. Mirrors assertEndpointOS pattern: iterate lines,
// unmarshal into map[string]any, check path fields. Skips records where the field is
// absent, empty, or "." (relative root placeholder).
func assertWindowsPathShape(t *testing.T, ndjson []byte) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(string(ndjson)), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		for _, field := range []string{"project_path", "source_file"} {
			if v, ok := rec[field].(string); ok && v != "" && v != "." {
				if strings.Contains(v, "/") {
					t.Errorf("assertWindowsPathShape: line %d field %q = %q contains forward slash", i+1, field, v)
				}
				if len(v) < 2 || v[1] != ':' {
					t.Errorf("assertWindowsPathShape: line %d field %q = %q missing drive letter", i+1, field, v)
				}
			}
		}
	}
}

// assertWindowsEndpointUID asserts every record with an endpoint sub-object
// has an empty uid field. Only called on Windows. Mirrors assertEndpointOS pattern.
func assertWindowsEndpointUID(t *testing.T, ndjson []byte) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(string(ndjson)), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		ep, ok := rec["endpoint"].(map[string]any)
		if !ok {
			continue
		}
		if uid, exists := ep["uid"]; exists {
			if uidStr, _ := uid.(string); uidStr != "" {
				t.Errorf("assertWindowsEndpointUID: line %d: endpoint.uid = %q, want empty on Windows", i+1, uidStr)
			}
		}
	}
}

// assertParityRecordCoverage fails if any of the 5 expected ecosystem strings
// is absent from the (already-normalized) NDJSON stream.
// Coverage per-OS + the assertion running identically on all three runners is
// what establishes PTEST-01 parity. Exact byte-equality across OSes is
// intentionally not required (OS path strings differ by design).
func assertParityRecordCoverage(t *testing.T, ndjson []byte) {
	t.Helper()
	// The 5 distinct ecosystem strings emitted by the 8 fixture package managers.
	// Source: internal/model/model.go ecosystem constants.
	required := []string{"npm", "pypi", "go", "rubygems", "packagist"}
	found := map[string]bool{}

	lines := strings.Split(strings.TrimSpace(string(ndjson)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		if eco, ok := rec["ecosystem"].(string); ok && eco != "" {
			found[eco] = true
		}
	}

	for _, eco := range required {
		if !found[eco] {
			t.Errorf("assertParityRecordCoverage: ecosystem %q not found in pollen output on %s (found: %v)",
				eco, runtime.GOOS, found)
		}
	}
}
