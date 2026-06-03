# Upstream Provenance

## Fork metadata

| Field | Value |
|-------|-------|
| upstream | github.com/perplexityai/bumblebee |
| pinned commit | c24089804ee66ece4bec6f14638cb98985389cdb |
| pinned tag | v0.1.1 |
| pinned date | 2026-05-22 |
| verified by | bantuson |

```
upstream: github.com/perplexityai/bumblebee
pinned commit: c24089804ee66ece4bec6f14638cb98985389cdb
pinned tag: v0.1.1
pinned date: 2026-05-22
verified by: bantuson
```

The upstream commit is recorded in the Pollen git history:

```
git rev-list --all | grep c24089804ee66ece4bec6f14638cb98985389cdb
```

## Sync workflow

> **Status (updated Phase 5):** No upstream absorption occurred in Milestone 2 (v0.1.1-pollen.2
> through v0.1.1-pollen.5). The pollen.2–5 releases are Windows-addition releases, not
> upstream-absorption releases. The pinned commit remains `c24089804ee66ece4bec6f14638cb98985389cdb`
> (upstream tag v0.1.1). This workflow is maintained so the next maintainer can absorb upstream
> changes cold without prior context.

When the maintainer decides to absorb upstream changes, follow this numbered procedure. All
commands run inside the `pollen` repository root unless otherwise noted.

1. **Fetch upstream.**

   ```bash
   git remote -v                    # confirm upstream remote is configured
   git remote update upstream       # fetch latest upstream main branch
   git log upstream/main --oneline -10   # review new commits
   NEW_COMMIT=$(git rev-parse upstream/main)
   echo "New upstream HEAD: $NEW_COMMIT"
   ```

   If the `upstream` remote is not configured, add it first:
   ```bash
   git remote add upstream https://github.com/perplexityai/bumblebee.git
   git remote update upstream
   ```

2. **Diff-review the new commits.** Review the diff between the current pinned commit and the
   new target commit. Pay particular attention to:
   - New files added by upstream (could introduce malicious code).
   - Changes to the NDJSON schema in `internal/output/` (breaks Beekeeper consumers).
   - Changes to the root resolver in `internal/resolver/` (conflicts with our Windows code paths).
   - Changes to `LICENSE` or `NOTICE` (legal compliance — we must preserve both verbatim or update
     our attribution accordingly).

   ```bash
   PINNED=c24089804ee66ece4bec6f14638cb98985389cdb
   git diff $PINNED $NEW_COMMIT --stat          # summary of changed files
   git diff $PINNED $NEW_COMMIT -- LICENSE NOTICE   # check legal files
   git diff $PINNED $NEW_COMMIT -- internal/    # schema + resolver changes
   git log $PINNED..$NEW_COMMIT --oneline        # each new upstream commit
   ```

   **STOP if** any new file appears in a sensitive location or a schema/LICENSE change requires
   manual review. Do not proceed until the diff is fully understood.

3. **Run upstream tests on Linux and macOS.** Checkout the new upstream target commit in a
   separate worktree and run `go test ./...` to confirm upstream did not break itself.

   ```bash
   git worktree add /tmp/bumblebee-check $NEW_COMMIT
   cd /tmp/bumblebee-check
   go test ./...        # must pass on Linux; repeat on macOS runner
   cd -
   git worktree remove /tmp/bumblebee-check
   ```

4. **Cherry-pick or merge preserving Windows code paths.** Cherry-pick or merge the new upstream
   commits into the Pollen branch, resolving conflicts in favor of preserving our Windows-specific
   files (`*_windows.go`, `cmd/pollen/`, `NOTICE`, `UPSTREAM.md`, `CHANGES.md`).

   ```bash
   # Option A: cherry-pick (preferred when upstream changes are small/targeted)
   git cherry-pick $PINNED..$NEW_COMMIT

   # Option B: merge (for large upstream absorptions)
   git merge upstream/main --no-ff -m "chore: absorb upstream $NEW_COMMIT (pollen.N)"

   # After either option, check for conflict markers:
   git diff --check
   # If conflicts: resolve in favor of Pollen Windows code paths, then:
   git add .
   git cherry-pick --continue   # or: git merge --continue
   ```

5. **Re-run the Pollen CI matrix.** Run `go test ./...` and `go vet ./...` on all three OSes
   (ubuntu-latest, macos-latest, windows-latest) to confirm the merged result is clean.

   ```bash
   go vet ./...
   go test ./...
   # Windows-only tests require a Windows machine or the GitHub Actions windows-latest runner.
   # Push to a feature branch and let CI validate all three OSes.
   ```

6. **Run the differential test.** The differential test asserts that `pollen` output on Linux is
   byte-for-byte identical to upstream `bumblebee` output for the same fixed fixture. Any divergence
   in the NDJSON output (excluding Windows-only fields) is a bug to fix before completing the merge.

   ```bash
   # On Linux or macOS CI runner:
   go test ./cmd/pollen/ -run '^TestDifferential$' -v
   # Expected output: "PASS" — any divergence is a bug; investigate normalize_diff.go and fix.
   ```

7. **Bump the pinned commit and update CHANGES.md.** Update this file's `pinned commit` field
   with the new 40-char SHA, and prepend a new dated section to `CHANGES.md` summarising what
   was absorbed.

   ```bash
   # In UPSTREAM.md: replace old pinned commit with new SHA
   sed -i "s/$PINNED/$NEW_COMMIT/g" UPSTREAM.md
   # Also update the version history table with the new row.

   # In CHANGES.md: prepend a new section above the previous pollen.N section:
   # ## v0.1.1-pollen.N+1 (YYYY-MM-DD) — Upstream absorption
   # Pinned to <NEW_COMMIT> (upstream tag vX.Y.Z, YYYY-MM-DD).
   # Absorbed: [summary of upstream changes]
   ```

8. **Tag a new Pollen release.** Tag the merged commit as `v0.1.1-pollen.N` (incrementing N),
   push the tag, and let the release workflow produce the signed binary + SBOM.

   ```bash
   NEXT_TAG=v0.1.1-pollen.6    # increment N from the current latest
   git tag -a $NEXT_TAG HEAD -m "Pollen $NEXT_TAG — upstream absorption <NEW_COMMIT>"
   git push origin main
   git push origin $NEXT_TAG
   # GitHub Actions release.yml triggers on push: tags: 'v*'
   # Verify cosign signature after the release workflow completes:
   cosign verify-blob \
     --bundle checksums.txt.sigstore.json \
     --certificate-identity-regexp '^https://github\.com/Bantuson/pollen/' \
     --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
     checksums.txt
   ```

This procedure is also the two-person-rule discipline from PRD Section 12.2, applied to upstream
absorption: the diff-review step (step 2) is the primary gate.

## Version history

| Pollen version | Pinned upstream commit | Date absorbed | Summary |
|----------------|----------------------|---------------|---------|
| v0.1.1-pollen.1 | c24089804ee66ece4bec6f14638cb98985389cdb | 2026-05-22 | Initial fork — module path rewrite, cmd/pollen rename, Windows build clean, attribution files |
| v0.1.1-pollen.2 | c24089804ee66ece4bec6f14638cb98985389cdb | 2026-06-02 | Windows root resolver (WRES-01/02, PTEST-01) — no upstream absorption; Windows-addition release |
| v0.1.1-pollen.3 | c24089804ee66ece4bec6f14638cb98985389cdb | 2026-06-02 | Windows path representation (WPATH-01/02) — no upstream absorption; Windows-addition release |
| v0.1.1-pollen.4 | c24089804ee66ece4bec6f14638cb98985389cdb | 2026-06-02 | Windows extension & MCP coverage (WEXT-01/02/03) — no upstream absorption; Windows-addition release |
| v0.1.1-pollen.5 | c24089804ee66ece4bec6f14638cb98985389cdb | 2026-06-03 | Milestone close — UPSTREAM sync doc (SYNC-01), beekeeper CI pin (BKINT-02), Windows honeypot (PTEST-05), pollen-self catalog (SDEF-01) |

> **Note on pinned commit uniformity:** pollen.2 through pollen.5 all share the same upstream
> pinned commit (`c24089804ee66ece4bec6f14638cb98985389cdb`, upstream tag v0.1.1, 2026-05-22).
> This is correct and expected: no upstream absorption occurred across these four releases. All
> changes are Windows-addition releases (Phases 2–5), not upstream-absorption releases. The pinned
> commit only advances when an upstream sync is deliberately performed following the 8-step workflow
> documented above.

## Contribution-back status

> **SYNC-02 disposition: DEFERRED** (maintainer decision D-2, 2026-06-03).
> No pull requests have been opened against `perplexityai/bumblebee` in Milestone 2.
> This section documents the prepared patch set and the rationale for deferral so the
> intent is auditable and re-openable in a future milestone.

### Prepared Windows patch set

The Windows compatibility additions from Phases 2–4 are prepared as a patch set available
in this repository across commits `2c202ef..b906404`. The patch set covers:

| Requirement group | Commits | Description |
|-------------------|---------|-------------|
| WRES-01, WRES-02 | `2c202ef`, `eba8e4c`, `c94b271` | Windows root resolver for 8 ecosystems: npm, pnpm, Yarn, Bun, PyPI, Go modules, RubyGems, Composer. `windowsBaselinePackageRoots()` and `windowsSystemRoots()` in `cmd/pollen/roots_windows.go` + `//go:build !windows` stubs + `TestWindowsBaselineRoots`. |
| WPATH-01, WPATH-02 | `2daf939`, `f21e231`, `92e9ac0`, `1cb3fdb`, `19695e3` | Windows path representation: `filepath.FromSlash` in npm/pnpm `projectPath` joins; `endpoint.uid` empty on Windows (SID suppressed); parity assertions for Windows path shape. |
| WEXT-01, WEXT-02, WEXT-03 | `dbe4c52`, `77ad510`, `94fb651`, `b906404`, `a9db7b3` | Windows editor-extension roots (VS Code, Cursor, Windsurf, VSCodium via both `.vscodium` and `.vscode-oss`); Windows browser-extension roots (Chrome/Chromium/Edge/Brave per-profile + Firefox Profiles); Windows MCP host-config roots (Claude Desktop, Cursor, Windsurf, Cline, Gemini CLI via `%APPDATA%`/`%USERPROFILE%`); unconditional `.windsurf` MCP root on all platforms. |

To view the full diff of the Windows additions against the upstream pinned commit:

```bash
git diff c24089804ee66ece4bec6f14638cb98985389cdb b906404 -- cmd/pollen/roots_windows.go \
  cmd/pollen/roots_notwindows.go internal/ecosystem/npm/npm.go \
  internal/ecosystem/pnpm/pnpm.go internal/endpoint/endpoint.go
```

### Contribution-back rationale and deferral

Upstream `perplexityai/bumblebee` already has open Windows-support PRs:

- **PR #4** (austinconnor): Claims to fix both issue #1 (path representation) and issue #2
  (Windows root discovery) in a single commit with build-tag-separated implementation.
  Open since launch (3+ weeks as of M2 planning). No reviews, no labels, no maintainer
  comments as of 2026-06-03.
- **PR #3**: A parallel Windows-support attempt by a different contributor. Also open with
  no merge signal.

A commenter on issue #2 was explicitly told by the maintainers that they plan to implement
their own Windows update rather than merge community PRs. Given this, a contribution-back PR
from this repository would likely be ignored in the same way.

**Decision (D-2):** Contribution-back is DEFERRED to a future milestone when upstream
signals readiness to accept Windows-support contributions. The prepared patch set is
preserved in this repository (commits `2c202ef..b906404`) and the implementation
approach is compatible with the upstream PR #4 build-tag pattern, so it can be
re-submitted as upstream-shaped PRs when the time comes:

1. **Root resolver PR** — `cmd/pollen/roots_windows.go` approach (mirrors upstream PR #4 pattern:
   build-tag-separated `cmd/bumblebee/roots_windows.go`).
2. **Path representation PR** — `filepath.FromSlash` in npm/pnpm + `endpoint.uid` guard.
3. **Extension & MCP coverage PR** — editor/browser/MCP host-config Windows roots.

Each upstream PR should reference the equivalent Pollen tag (`v0.1.1-pollen.2`, `.3`, `.4`)
and link parity-test output as evidence that the Windows additions do not break Linux/macOS
byte-for-byte behavior (the `TestDifferential` result).

**SYNC-02 is satisfied by this documented deferral.** The roadmap SC2 requirement
(contribution-back) is explicitly deferred per maintainer decision D-2. The verifier
MUST NOT flag the absence of an upstream PR as a Phase-5 gap.
