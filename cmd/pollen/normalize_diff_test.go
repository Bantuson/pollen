package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestNormalize verifies the NDJSON normalization harness used by the
// differential test (PTEST-02). These tests MUST pass before the differential
// guard is meaningful — they verify that normalize() removes exactly the
// right fields and produces stable output.
func TestNormalize(t *testing.T) {
	t.Run("strips 7 non-deterministic fields and preserves the rest", func(t *testing.T) {
		// A synthetic record with all 7 non-det fields plus stable fields.
		input := `{"record_type":"package","record_id":"package:abc123","schema_version":"0.1.0","scanner_name":"pollen","scanner_version":"0.1.1","run_id":"deadbeef","scan_time":"2026-06-01T10:00:00Z","endpoint":{"hostname":"myhost","os":"linux","arch":"amd64","username":"alice","uid":"1000","device_id":"dev-42"},"profile":"baseline","ecosystem":"npm","package_name":"leftpad","normalized_name":"leftpad","version":"1.0.0","source_type":"lockfile","source_file":"/home/alice/project/package-lock.json","has_lifecycle_scripts":false,"confidence":"high"}`

		out, err := normalize([]byte(input))
		if err != nil {
			t.Fatalf("normalize() error = %v", err)
		}

		// Decode the single output line to inspect.
		var got map[string]any
		if err := json.Unmarshal(out, &got); err != nil {
			t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
		}

		// Fields that MUST be absent after normalization.
		for _, key := range []string{"run_id", "scan_time", "end_time", "duration_ms"} {
			if _, exists := got[key]; exists {
				t.Errorf("normalize() left key %q in output; should have stripped it", key)
			}
		}

		// Endpoint sub-fields that MUST be absent.
		if ep, ok := got["endpoint"].(map[string]any); ok {
			for _, key := range []string{"hostname", "username", "uid"} {
				if _, exists := ep[key]; exists {
					t.Errorf("normalize() left endpoint.%s in output; should have stripped it", key)
				}
			}
			// Non-stripped endpoint fields should remain.
			for _, key := range []string{"os", "arch"} {
				if _, exists := ep[key]; !exists {
					t.Errorf("normalize() stripped endpoint.%s; should have preserved it", key)
				}
			}
		} else {
			t.Error("normalize() removed the 'endpoint' key entirely; expected partial endpoint to remain")
		}

		// Stable fields that MUST be preserved.
		for _, key := range []string{
			"record_id", "record_type", "schema_version", "scanner_name",
			"ecosystem", "package_name", "version",
		} {
			if _, exists := got[key]; !exists {
				t.Errorf("normalize() stripped key %q; should have preserved it", key)
			}
		}
	})

	t.Run("sorts 3-record stream in reverse record_id order to ascending order", func(t *testing.T) {
		// Three records with record_ids that spell "c" > "b" > "a" when sorted ascending.
		r1 := `{"record_type":"package","record_id":"package:ccc","schema_version":"0.1.0","scanner_name":"pollen","run_id":"r1","scan_time":"2026-01-01T00:00:00Z","endpoint":{}}`
		r2 := `{"record_type":"package","record_id":"package:bbb","schema_version":"0.1.0","scanner_name":"pollen","run_id":"r2","scan_time":"2026-01-01T00:00:00Z","endpoint":{}}`
		r3 := `{"record_type":"package","record_id":"package:aaa","schema_version":"0.1.0","scanner_name":"pollen","run_id":"r3","scan_time":"2026-01-01T00:00:00Z","endpoint":{}}`
		// Input in reverse order (c, b, a); expected output order is a, b, c.
		input := strings.Join([]string{r1, r2, r3}, "\n")

		out, err := normalize([]byte(input))
		if err != nil {
			t.Fatalf("normalize() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) != 3 {
			t.Fatalf("normalize() returned %d lines, want 3", len(lines))
		}

		wantIDs := []string{"package:aaa", "package:bbb", "package:ccc"}
		for i, line := range lines {
			var rec map[string]any
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				t.Fatalf("line %d is not valid JSON: %v", i, err)
			}
			gotID, _ := rec["record_id"].(string)
			if gotID != wantIDs[i] {
				t.Errorf("line %d record_id = %q, want %q", i, gotID, wantIDs[i])
			}
		}
	})

	t.Run("two streams differing only in stripped fields normalize to byte-identical output", func(t *testing.T) {
		// Simulates pollen vs bumblebee output: same package, different run_id/scan_time/endpoint.
		// This is the core PTEST-02 invariant.
		pollenStream := `{"record_type":"package","record_id":"package:stable123","schema_version":"0.1.0","scanner_name":"pollen","run_id":"pollen-run-aabbcc","scan_time":"2026-06-01T10:00:00Z","endpoint":{"hostname":"pollen-host","os":"linux","arch":"amd64","username":"pollenbee","uid":"1001"},"profile":"baseline","ecosystem":"npm","package_name":"express","normalized_name":"express","version":"4.18.2","source_type":"lockfile","source_file":"/project/package-lock.json","has_lifecycle_scripts":false,"confidence":"high"}`
		bumblebeeStream := `{"record_type":"package","record_id":"package:stable123","schema_version":"0.1.0","scanner_name":"bumblebee","run_id":"bee-run-112233","scan_time":"2025-12-01T08:30:00Z","endpoint":{"hostname":"bee-host","os":"linux","arch":"amd64","username":"beehive","uid":"2002"},"profile":"baseline","ecosystem":"npm","package_name":"express","normalized_name":"express","version":"4.18.2","source_type":"lockfile","source_file":"/project/package-lock.json","has_lifecycle_scripts":false,"confidence":"high"}`

		pollenOut, err := normalize([]byte(pollenStream))
		if err != nil {
			t.Fatalf("normalize(pollen) error = %v", err)
		}
		bumblebeeOut, err := normalize([]byte(bumblebeeStream))
		if err != nil {
			t.Fatalf("normalize(bumblebee) error = %v", err)
		}

		if !bytes.Equal(pollenOut, bumblebeeOut) {
			t.Errorf("normalize() outputs differ:\npollen:    %s\nbumblebee: %s", pollenOut, bumblebeeOut)
		}
	})

	t.Run("record missing record_id returns an error (fail-closed on bad input)", func(t *testing.T) {
		// A record that lacks the record_id field entirely.
		input := `{"record_type":"package","schema_version":"0.1.0","scanner_name":"pollen","run_id":"x","scan_time":"2026-01-01T00:00:00Z","endpoint":{},"ecosystem":"npm"}`

		_, err := normalize([]byte(input))
		if err == nil {
			t.Fatal("normalize() returned nil error for record missing record_id; want a non-nil error (fail-closed)")
		}
	})

	t.Run("empty record_id string returns an error (fail-closed on bad input)", func(t *testing.T) {
		// A record where record_id is present but empty string.
		input := `{"record_type":"package","record_id":"","schema_version":"0.1.0","scanner_name":"pollen","run_id":"x","scan_time":"2026-01-01T00:00:00Z","endpoint":{}}`

		_, err := normalize([]byte(input))
		if err == nil {
			t.Fatal("normalize() returned nil error for record with empty record_id; want a non-nil error (fail-closed)")
		}
	})

	t.Run("blank lines in NDJSON stream are skipped gracefully", func(t *testing.T) {
		r1 := `{"record_type":"package","record_id":"package:zzz","schema_version":"0.1.0","run_id":"r","scan_time":"t","endpoint":{}}`
		// Two records separated by a blank line (common in some emitters).
		input := r1 + "\n\n" + r1

		out, err := normalize([]byte(input))
		if err != nil {
			t.Fatalf("normalize() with blank lines error = %v", err)
		}
		// Two identical record_ids: after sort both should still appear (dedup is NOT part of normalize).
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) != 2 {
			t.Fatalf("normalize() returned %d lines, want 2 (blank lines should be skipped, duplicates kept)", len(lines))
		}
	})
}
