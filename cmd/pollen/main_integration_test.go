package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUsage(t *testing.T) {
	var buf bytes.Buffer
	usage(&buf)
	if !strings.Contains(buf.String(), "pollen") || !strings.Contains(buf.String(), "scan") {
		t.Errorf("usage() output missing expected text:\n%s", buf.String())
	}
}

func TestRunRoots(t *testing.T) {
	if code := runRoots([]string{"--profile", "baseline"}); code != 0 {
		t.Errorf("runRoots baseline exit = %d, want 0", code)
	}
	if code := runRoots([]string{"--profile", "bogus"}); code == 0 {
		t.Error("runRoots with bogus profile should return non-zero")
	}
}

// TestRunScanFileOutput drives the scan subcommand end-to-end over a fixture
// tree, writing NDJSON to a file, and asserts a scan_summary terminator was
// emitted (always-on by default), plus the npm lockfile record.
func TestRunScanFileOutput(t *testing.T) {
	proj := t.TempDir()
	// A minimal npm v3 lockfile so the npm scanner emits a package record.
	lock := `{"name":"fixture","version":"1.0.0","lockfileVersion":3,` +
		`"packages":{"":{"name":"fixture"},"node_modules/left-pad":{"version":"1.3.0"}}}`
	if err := os.WriteFile(filepath.Join(proj, "package-lock.json"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "records.ndjson")
	code := runScan([]string{
		"--profile", "project", "--root", proj,
		"--output", "file", "--output-file", out,
	})
	if code != 0 {
		t.Fatalf("runScan exit = %d, want 0", code)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "scan_summary") {
		t.Errorf("output missing scan_summary record:\n%s", body)
	}
	if !strings.Contains(body, "left-pad") {
		t.Errorf("output missing the left-pad package record:\n%s", body)
	}
}

func TestRunScanErrors(t *testing.T) {
	// Unknown profile.
	if code := runScan([]string{"--profile", "bogus", "--root", t.TempDir()}); code == 0 {
		t.Error("runScan with unknown profile should return non-zero")
	}
	// deep requires an explicit --root.
	if code := runScan([]string{"--profile", "deep"}); code == 0 {
		t.Error("runScan deep without --root should return non-zero")
	}
}
