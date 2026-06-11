# Changes from upstream

This file documents every significant change Pollen makes from the pinned upstream commit
`c24089804ee66ece4bec6f14638cb98985389cdb` (tag v0.1.1, 2026-05-22) of
`github.com/perplexityai/bumblebee`. Each sync absorption appends a new dated section.

---

## v0.2.0 (2026-06-11) — First standalone public release

First public release of Pollen as a standalone, Windows-focused inventory
scanner. It consolidates all the Windows-support work previously staged across
the (never-tagged) `v0.1.1-pollen.2`–`.5` planning sections below into a single
clean semver release, and drops the `-pollen.N` fork-suffix scheme.

`schema_version` stays `0.1.0` (behavioral fork, not a protocol fork); the
Linux/macOS differential (`TestDifferential`) remains byte-for-byte identical to
upstream Bumblebee.

### Included since v0.1.1-pollen.1

- **Windows root resolver** (WRES-01/02): per-user + system package roots for
  npm, pnpm, Yarn, Bun, PyPI, Go modules, RubyGems, Composer.
- **Windows path representation** (WPATH-01/02): `filepath.FromSlash` project-path
  joins; `endpoint.uid` empty on Windows (SID suppressed).
- **Windows editor/browser/MCP coverage** (WEXT-01/02/03): VS Code / Cursor /
  Windsurf / VSCodium extension roots; Chromium-family + Firefox browser-extension
  roots; Claude Desktop / Cursor / Windsurf / Cline / Gemini MCP host-config roots.
- **Test coverage hardening**: model 100%, Windows-compatible walk coverage,
  scanner error classifiers, editorext 98.5%, cmd/pollen integration tests, plus
  an `internal/coveragegate` presence gate with a reason-coded no-test allowlist.
- **License / attribution**: Apache-2.0 `LICENSE`, `NOTICE` with an explicit
  Apache-2.0 §4(b) modification statement, and `UPSTREAM.md` provenance.

The original per-phase change log is retained below for provenance; those
`pollen.N` versions were prepared during the upstream v1.1.0 milestone but never
tagged.

---

## v0.1.1-pollen.5 (2026-06-03) — Milestone close

> **Status: prepared, not yet tagged.** The version bump and this delta are committed locally;
> the `v0.1.1-pollen.5` git tag + Sigstore signing + CycloneDX SBOM (via `release.yml`) are
> **deferred to milestone (M2) close** by maintainer decision (D-05). The `v0.1.1-pollen.2`,
> `v0.1.1-pollen.3`, `v0.1.1-pollen.4`, and `v0.1.1-pollen.5` signed-release obligations are
> all batched together at M2 close and cut in order (pollen.2 → .3 → .4 → .5) by the maintainer.
> See beekeeper STATE.md Deferred Items + the Phase 5 release runbook for the exact tag/verify
> commands.

This is the milestone-complete tag for Beekeeper v1.1.0 "Pollen". It closes the milestone
and marks the transition from local-only to publicly-signed releases.

No Pollen **source code** changed in this release — all milestone-close work landed in the
beekeeper repo. This release provides the metadata layer (UPSTREAM.md sync doc, VERSION bump,
CHANGES.md entry) required before the batched tag cut.

`schema_version` stays `0.1.0` (behavioral fork, not protocol fork). The Linux/macOS
differential (`TestDifferential`) is unaffected — no detection-logic changes in this release.

### Added / Updated (metadata only)

- `UPSTREAM.md` — Extended with concrete 8-step sync workflow (runnable commands, worked
  example invocations), version-history table rows for pollen.2/3/4/5, and the
  `## Contribution-back status` section documenting the prepared Windows patch set
  (commits `2c202ef..b906404`), the upstream PR #3/#4 context, and the DEFERRED
  disposition per maintainer decision D-2 (SYNC-01 satisfied; SYNC-02 satisfied-by-documented-
  deferral). A second maintainer can now follow the upstream sync workflow cold.
- `VERSION` — Bumped from `0.1.1-pollen.4` to `0.1.1-pollen.5`.
- `CHANGES.md` — This entry.

### Cross-repo deliverables (beekeeper repo)

The following beekeeper deliverables close the milestone alongside this Pollen release; they
are tracked in beekeeper's STATE.md and ROADMAP.md:

- **BKINT-02** — beekeeper CI installs Pollen at a pinned version (`go install
  github.com/bantuson/pollen/cmd/pollen@v0.1.1-pollen.5`); Windows inventory-test skip
  baseline is zero after this phase.
- **PTEST-05** — Windows Sentry honeypot E2E: planted process tree reads synthetic
  `%USERPROFILE%\.aws\credentials` and makes outbound connection; beekeeper's
  `exfil-signature-fusion` rule (SENTRY-005) fires on the Windows CI runner.
- **SDEF-01** — `pollen-self` entries added to the unified `beekeeper-self` catalog;
  `beekeeper selftest` passes with the extended catalog.

### Contribution-back (SYNC-02 — deferred)

No upstream PRs were opened against `perplexityai/bumblebee` in Milestone 2. The Windows
patch set is prepared and preserved in commits `2c202ef..b906404`. Contribution-back is
deferred to a future milestone per maintainer decision D-2. See `UPSTREAM.md §Contribution-back
status` for the full rationale and re-submission guide.

---

## v0.1.1-pollen.4 (2026-06-02) — Windows extension & MCP coverage

> **Status: prepared, not yet tagged.** The version bump and this delta are committed locally;
> the `v0.1.1-pollen.4` git tag + Sigstore signing + CycloneDX SBOM (via `release.yml`) are
> **deferred to milestone (M2) close** by maintainer decision (D-06). The `v0.1.1-pollen.2`,
> `v0.1.1-pollen.3`, and `v0.1.1-pollen.4` signed-release obligations are all batched together
> at M2 close. See beekeeper STATE.md Deferred Items for the pending tag/verify commands.

Satisfies WEXT-01 (five editor-extension roots including VSCodium via both `.vscodium` and
`.vscode-oss`), WEXT-02 (Chrome/Chromium/Edge/Brave per-profile `Extensions/` directories +
Firefox `Profiles` parent on Windows), and WEXT-03 (Claude Desktop + Cline via `%APPDATA%`,
Cursor/Windsurf/Gemini via `%USERPROFILE%` dotfiles, plus the new unconditional `.windsurf`
MCP root fixing the Windsurf `mcp.json` gap on all platforms). `schema_version` stays `0.1.0`
(behavioral fork, not protocol fork). The Linux/macOS differential (`TestDifferential`) remains
byte-identical — all WEXT changes are Windows-only code paths; no Unix bytes changed.

### Modified

- `cmd/pollen/roots.go` — filled the two empty `case "windows":` skeletons in
  `browserExtensionCandidateRoots`: Chrome, Chromium, Edge, Brave under `%LOCALAPPDATA%` (per-profile
  `Extensions/` dirs) and Firefox `%APPDATA%\Mozilla\Firefox\Profiles` parent (using per-variable
  env-var guards per Phase-2 Pitfall-5 discipline). Added `.vscode-oss/extensions` to
  `baselineHomeCandidates` (alongside `.vscodium/extensions`) so both VSCodium install variants
  resolve (WEXT-01). Added `filepath.Join(home, ".windsurf")` `RootKindMCPConfig` root
  unconditionally before the `runtime.GOOS` switch so `%USERPROFILE%\.windsurf\mcp.json` is
  reachable on all platforms (fixes the `.windsurf` MCP gap, WEXT-03). Added `case "windows":` MCP
  block with `%APPDATA%\Claude` and `%APPDATA%\cline` roots (Cursor/Windsurf/Gemini already covered
  by cross-platform dotfile roots).
- `internal/ecosystem/editorext/editorext.go` — added `.vscode-oss/extensions` to
  `extensionRootSegments` (alongside `.vscodium/extensions`), covering the alternate VSCodium
  install path prescribed by PRD §8.2 (WEXT-01).

### Added

- `cmd/pollen/roots_windows_test.go` — `TestWindowsExtensionMCPRoots`: Windows-only fixture test
  (`//go:build windows`, zero `t.Skip` calls) asserting all five editor-extension roots
  (VS Code, VS Code Insiders, Cursor, Windsurf, VSCodium via both `.vscodium` and `.vscode-oss`),
  all Chromium-family browser roots (Chrome, Chromium, Edge, Brave per-profile) + Firefox Profiles
  parent, and all five MCP host-config roots (Claude Desktop, Cursor, Windsurf, Cline, Gemini CLI).
  Test isolation via `t.Setenv(USERPROFILE/APPDATA/LOCALAPPDATA/ProgramFiles)` — never `HOME`.

---

## v0.1.1-pollen.3 (2026-06-02) — Windows path representation

> **Status: prepared, not yet tagged.** The version bump and this delta are committed locally;
> the `v0.1.1-pollen.3` git tag + Sigstore signing + CycloneDX SBOM (via `release.yml`) are
> **deferred to milestone (M2) close** by maintainer decision (D-06). Both the `v0.1.1-pollen.2`
> and `v0.1.1-pollen.3` signed-release obligations are batched together at M2 close. See beekeeper
> STATE.md Deferred Items for the pending tag/verify commands.

Satisfies WPATH-01 (Windows project_path / source_file use backslash separators + drive letter)
and WPATH-02 (endpoint.uid is empty on Windows — SID strings suppressed). schema_version stays
`0.1.0` (behavioral fork, not protocol fork). The Linux/macOS differential (`TestDifferential`)
remains byte-identical — the WPATH changes are Windows-only code paths; no Unix bytes changed.

### Modified

- `internal/ecosystem/npm/npm.go` — `IsNodeModulesPackageJSON` wraps the `projectPath` join in
  `filepath.FromSlash`: no-op on Linux/macOS (forward slash is native), converts to backslash on
  Windows. Fixes WPATH-01 forward-slash leak in npm node_modules project_path.
- `internal/ecosystem/pnpm/pnpm.go` — `IsPnpmStorePackageJSON` applies the same
  `filepath.FromSlash` wrap to its `projectPath` join. Symmetric fix to npm.go (WPATH-01).
- `internal/endpoint/endpoint.go` — `UID` assignments in `Current()` guarded by
  `runtime.GOOS != "windows"`: both the `user.Current().Uid` happy path and the
  `os.Getuid()` fallback are skipped on Windows, leaving `endpoint.uid` as the zero-value empty
  string. On Linux/macOS the numeric UID is unchanged (D-04 regression guard). Fixes WPATH-02.
- `cmd/pollen/parity_test.go` — `TestParityAllEcosystems` gains a Windows-only block
  (after `assertEndpointOS`, before `normalize`) calling `assertWindowsPathShape` (every
  non-empty project_path / source_file has a drive letter and no forward slash) and
  `assertWindowsEndpointUID` (every endpoint sub-object has empty uid). On Linux/macOS the
  block is skipped; the existing parity assertions run unchanged.

### Added

- `internal/ecosystem/npm/npm_test.go` — Windows-gated unit tests
  `TestIsNodeModulesPackageJSONWindowsPath` and `TestIsNodeModulesPackageJSONScopedWindowsPath`
  (skip on non-Windows via `runtime.GOOS != "windows"` guard; assert backslash drive-letter
  projectPath for flat and scoped npm packages).
- `internal/ecosystem/pnpm/pnpm_test.go` — Windows-gated unit test
  `TestIsPnpmStorePackageJSONWindowsPath` (skip on non-Windows; asserts backslash drive-letter
  projectPath for pnpm store layout).
- `internal/endpoint/endpoint_test.go` — `TestCurrentWindowsUID`: on Windows asserts
  `endpoint.uid == ""`; on Linux/macOS asserts uid is non-empty numeric string (D-04 regression).

---

## v0.1.1-pollen.2 (2026-06-02) — Windows root resolver

> **Status: prepared, not yet tagged.** The version bump and this delta are committed locally;
> the `v0.1.1-pollen.2` git tag + Sigstore signing + CycloneDX SBOM (via `release.yml`) are
> **deferred to milestone (M2) close** by maintainer decision. See beekeeper STATE.md / ROADMAP
> Phase 2 for the pending-release obligation and the exact tag/verify commands.

Adds Windows package-root discovery for all eight ecosystems (Phase 2 / PRD §8.1 M2.2),
satisfying WRES-01, WRES-02, and PTEST-01. All Windows code is build-tag-isolated so the
Linux/macOS bytes — and the `TestDifferential` byte-for-byte guard — are unchanged.

### Added

- `cmd/pollen/roots_windows.go` (`//go:build windows`) — `windowsBaselinePackageRoots()` and
  `windowsSystemRoots()`: Windows root discovery for npm, pnpm, Yarn, Bun, PyPI, Go modules,
  RubyGems, and Composer, built from `%APPDATA%` / `%LOCALAPPDATA%` / `%USERPROFILE%` /
  `%ProgramFiles%` via `filepath.Join` + `os.Getenv` (every env-var read guarded against empty),
  with `globExisting` for the `Python*`, `.gem/ruby/*`, and `Ruby*` wildcard roots.
- `cmd/pollen/roots_notwindows.go` (`//go:build !windows`) — nil stubs of both functions so the
  `case "windows":` bodies in `roots.go` compile on Linux/macOS (Go compiles all `runtime.GOOS`
  switch case bodies regardless of platform).
- `cmd/pollen/roots_windows_test.go` (`//go:build windows`) — `TestWindowsBaselineRoots`
  (8-ecosystem coverage) + empty-`%APPDATA%` guard test.
- `cmd/pollen/parity_test.go` — `TestParityAllEcosystems` (PTEST-01): runs on all three OSes,
  asserts `endpoint.os == runtime.GOOS` and equivalent normalized records across platforms.
  Reuses `buildCurrentPollen` / `runBinaryOnFixture` / `normalize` from the differential harness;
  `normalize_diff.go` is unchanged.
- `cmd/pollen/testdata/parity-fixture/` — single fake-package fixture tree consumed by all three
  OS parity runs (npm, pnpm, Yarn, Bun, PyPI, Go, RubyGems, Composer canary packages).

### Modified

- `cmd/pollen/roots.go` — added `case "windows":` delegation in `baselineHomeCandidates`,
  `systemRoots`, and both `browserExtensionCandidateRoots` switches (the browser case is an
  intentional empty Phase-4 skeleton); `isBroadHomeRoot` now refuses Windows drive roots (`C:\`),
  the `%USERPROFILE%` parent (`C:\Users`), and immediate user homes (`C:\Users\<name>`) via
  `filepath.VolumeName` — a no-op off Windows.
- `cmd/pollen/main_test.go` — flipped the 6 Phase-2 `t.Skip` markers (Windows root-resolver
  support is now implemented) and added Windows cases to `TestIsBroadHomeRoot`. The
  `TestDifferential` Windows skip stays: the differential is Linux+macOS-only by design.

### Removed

- (none)

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
- **`scanner_name` NDJSON field:** `"bumblebee"` → `"pollen"` (`internal/model/model.go`
  `ScannerName` constant). Every emitted NDJSON record now carries `"scanner_name": "pollen"`.
  Upstream continues to emit `"scanner_name": "bumblebee"`. This is an **intentional, documented
  fork divergence** (FORK-04 trademark + honest self-identification). The differential test
  (`TestDifferential` in `cmd/pollen/differential_test.go`) normalizes this field out of both
  sides before the byte-for-byte assertion so the test proves DETECTION-LOGIC parity, not
  self-identification-string parity. The normalization carve-out is documented in
  `cmd/pollen/normalize_diff.go`.
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

### Fixed (release bring-up, 2026-06-02)

These surfaced only when the first GitHub Actions CI + release pipeline ran — none could be
caught on the Windows dev box, where `TestDifferential` skips and the release pipeline cannot
run locally:

- **Differential normalization also strips `scanner_version`** (`cmd/pollen/normalize_diff.go`).
  pollen emits its own Go build/VCS version (`v0.0.0-<pseudo>` untagged, `v0.1.1-pollen.1` tagged)
  while upstream's tagged clone emits `v0.1.1`. Like `scanner_name`, this is build identity, not
  detection logic; stripping it keeps `TestDifferential` a DETECTION-LOGIC parity check. With this,
  the differential is green on Linux+macOS CI.
- **CI `go.mod/go.sum` tidiness check made zero-dependency-safe** (`.github/workflows/ci.yml`):
  pollen has no external deps, so `go.sum` does not exist; the check now runs under `shell: bash`
  and only asserts `go.sum` when present.
- **`.goreleaser.yaml` corrected for GoReleaser v2:** top-level `checksums:` → `checksum:` (v2
  schema), and added `main: ./cmd/pollen` (the build entry otherwise defaulted to the repo root,
  which has no `main`). Validated with `goreleaser check` + `goreleaser build --snapshot`.
- **cosign identity casing (Assumption A1 resolved):** the real Sigstore certificate subject is
  `https://github.com/Bantuson/pollen/.github/workflows/release.yml@refs/tags/v0.1.1-pollen.1`
  — GitHub OIDC uses the canonical account casing `Bantuson` (capital B), not the lowercase Go
  module path. `docs/THREAT-MODEL.md`'s `--certificate-identity-regexp` corrected to
  `^https://github\.com/Bantuson/pollen/`; `cosign verify-blob` returns `Verified OK`.
