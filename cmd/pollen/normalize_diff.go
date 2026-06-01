package main

// normalize_diff.go — NDJSON normalization harness for the differential test.
//
// Purpose: the differential test (PTEST-02) asserts that pollen and upstream
// bumblebee produce byte-for-byte identical NDJSON on a fixed fixture AFTER
// normalization. Raw output always differs because of:
//   - 7 non-deterministic fields (run_id, scan_time, end_time, duration_ms,
//     endpoint.hostname, endpoint.username, endpoint.uid)
//   - scanner_name field — intentional fork divergence: pollen emits "pollen",
//     upstream bumblebee emits "bumblebee". This is a documented fork delta
//     (CHANGES.md "Modified") that proves honest self-identification, not
//     behavioral drift. normalize() strips scanner_name from both sides so the
//     differential asserts DETECTION-LOGIC parity, not self-identification parity.
//   - record ordering — the scanner spawns 4 concurrent workers; records emit
//     in worker-completion order (non-deterministic). Sorting by record_id
//     (a stable SHA-256 content key over the package identity tuple) removes
//     this variance.
//
// normalize() strips the documented fields and sorts the resulting records by
// record_id. It fails closed (returns a non-nil error) if any record lacks a
// record_id or has an empty record_id, rather than silently sorting to an
// arbitrary position. This fail-closed behavior means a truncated or malformed
// feed is surfaced as a test failure, not a false pass.
//
// Fields stripped (8 total = 7 non-deterministic + 1 documented fork divergence):
//
//	Top-level (5):  run_id, scan_time, end_time, duration_ms, scanner_name
//	Endpoint (3):   hostname, username, uid
//
// Sort key: record_id (string, ascending). record_id is a SHA-256 of the
// record's content identity tuple (see internal/model/model.go StableID()).
// Same input package identity → same hash → stable deterministic sort anchor.

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// topLevelStripKeys are the top-level JSON keys deleted from every record
// before comparison. These fields are either non-deterministic across runs or
// are documented fork divergences between pollen and upstream bumblebee.
//
// Non-deterministic (4):
//   - run_id       — crypto/rand hex, different every run
//   - scan_time    — wall-clock RFC3339Nano, different every run
//   - end_time     — wall-clock (scan_summary only)
//   - duration_ms  — elapsed wall-clock time (scan_summary only)
//
// Documented fork divergence (1):
//   - scanner_name — pollen emits "pollen"; upstream emits "bumblebee".
//     This is intentional (FORK-04 trademark + honest identity). Stripping it
//     here means the differential asserts detection-logic parity, not
//     self-identification-string parity. The carve-out is documented in
//     CHANGES.md under Modified.
var topLevelStripKeys = map[string]struct{}{
	"run_id":       {},
	"scan_time":    {},
	"end_time":     {},
	"duration_ms":  {},
	"scanner_name": {},
}

// endpointStripKeys are the endpoint sub-object JSON keys deleted from every
// record. These fields are machine-specific and differ across CI runners and
// developer machines.
//
//   - hostname — os.Hostname()
//   - username — user.Current().Username
//   - uid      — user.Current().Uid / os.Getuid()
var endpointStripKeys = map[string]struct{}{
	"hostname": {},
	"username": {},
	"uid":      {},
}

// normalize parses an NDJSON stream (one JSON object per non-blank line),
// strips the documented non-deterministic and fork-divergent fields, sorts the
// records ascending by record_id, and returns the result as newline-joined
// re-marshalled JSON.
//
// Go's json.Marshal produces deterministic output for map[string]any:
// map keys are sorted alphabetically in the output. This means any two records
// with identical content (after stripping) produce byte-identical marshal
// output regardless of the original field ordering in the source stream.
//
// Returns a non-nil error if:
//   - any non-blank line is not valid JSON
//   - any record lacks a "record_id" key or has an empty "record_id" value
//     (fail-closed: a malformed feed surfaces as a test failure, not a
//     false pass with undefined sort order)
func normalize(ndjson []byte) ([]byte, error) {
	lines := strings.Split(string(ndjson), "\n")

	var records []map[string]any

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, fmt.Errorf("normalize: line %d is not valid JSON: %w", i+1, err)
		}

		// Strip top-level non-deterministic / fork-divergent keys.
		for k := range topLevelStripKeys {
			delete(rec, k)
		}

		// Strip machine-specific keys inside the endpoint sub-object.
		// If the endpoint key exists and is a map, strip the three keys.
		// If endpoint is absent or not a map, leave it alone.
		if ep, ok := rec["endpoint"].(map[string]any); ok {
			for k := range endpointStripKeys {
				delete(ep, k)
			}
			// Re-assign the mutated sub-object (map is reference, but be explicit).
			rec["endpoint"] = ep
		}

		// Fail-closed: record_id must be a non-empty string.
		// Silently sorting a record without a record_id would produce
		// undefined comparison order and could mask real behavioral drift.
		rid, ok := rec["record_id"]
		if !ok {
			return nil, fmt.Errorf("normalize: line %d is missing required 'record_id' field (fail-closed: malformed record, not a silent pass)", i+1)
		}
		ridStr, ok := rid.(string)
		if !ok || ridStr == "" {
			return nil, fmt.Errorf("normalize: line %d has empty or non-string 'record_id' (fail-closed: malformed record, not a silent pass)", i+1)
		}

		records = append(records, rec)
	}

	// Sort ascending by record_id. record_id is a stable SHA-256 content key
	// (model.StableID()) so identical package observations always produce the
	// same record_id regardless of machine or run. Sorting by record_id removes
	// the worker-completion-order non-determinism from the comparison.
	sort.Slice(records, func(i, j int) bool {
		a, _ := records[i]["record_id"].(string)
		b, _ := records[j]["record_id"].(string)
		return a < b
	})

	// Re-marshal each record. json.Marshal sorts map keys alphabetically, so
	// two records with identical content always produce byte-identical output
	// regardless of the original field insertion order.
	var sb strings.Builder
	for idx, rec := range records {
		b, err := json.Marshal(rec)
		if err != nil {
			return nil, fmt.Errorf("normalize: marshal record %d: %w", idx, err)
		}
		if idx > 0 {
			sb.WriteByte('\n')
		}
		sb.Write(b)
	}

	return []byte(sb.String()), nil
}
