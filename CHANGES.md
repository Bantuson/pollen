# Changes from upstream

This file documents every significant change Pollen makes from the pinned upstream commit
`c24089804ee66ece4bec6f14638cb98985389cdb` (tag v0.1.1, 2026-05-22) of
`github.com/perplexityai/bumblebee`. Each sync absorption appends a new dated section.

---

## v0.1.1-pollen.1 (2026-06-01) — Initial fork

### Renamed

- `cmd/bumblebee/` → `cmd/pollen/` (binary name and directory)

### Modified

- **Module path:** `github.com/perplexityai/bumblebee` → `github.com/bantuson/pollen`
  (all `.go` imports updated)
- **selftest temp dir prefix:** `bumblebee-selftest-` → `pollen-selftest-`
  (`cmd/pollen/selftest.go` line 51)
- **version default:** `fileDefault` changed from `"0.1.1"` to `"0.1.1-pollen.1"`
  (`cmd/pollen/version.go`)
- **help / usage binary name:** references to `bumblebee` in flag help text and usage strings
  updated to `pollen`
- **Env var test overrides:** `BUMBLEBEE_USERS_DIR` → `POLLEN_USERS_DIR`,
  `BUMBLEBEE_TEST_DEVICE_ID` → `POLLEN_TEST_DEVICE_ID` (trademark discipline, FORK-04)
- **Test path separators:** `internal/scanner/scanner_test.go` `TestEndToEndScan` — hardcoded
  `/proj/` and `/dup/` path separators replaced with `filepath.Separator` for Windows portability
- **Test skips:** 6 Unix-specific test functions in `cmd/pollen/main_test.go` marked with
  structured `t.Skip` (Windows root-resolver support deferred to v0.1.1-pollen.2)

### Added

- `NOTICE` — Apache-2.0 attribution per PRD §7.2 (verbatim)
- `UPSTREAM.md` — Pinned upstream SHA + sync workflow (PRD §6.2)
- `CHANGES.md` — This file; delta log from the pinned commit
- `VERSION` — File containing `0.1.1-pollen.1`
- `threat_intel/README.md` — Notes that `threat_intel/` ships empty by design; catalogs flow
  through `beekeeper catalogs sync` (PRD §6.3 "reference" decision)

### Removed

- (none in initial fork)
