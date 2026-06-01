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

> **Status:** Sync EXECUTION is deferred to Phase 5 (v0.1.1-pollen.5). This document records the
> workflow now so that the Phase 5 executor has a fixed procedure to follow.

When the maintainer decides to absorb upstream changes, follow this numbered procedure:

1. **Fetch upstream.** Run `git remote update upstream` to fetch the latest upstream main branch.
   Confirm the remote is configured: `git remote -v`.

2. **Diff-review the new commits.** Review the diff between the current pinned commit and the new
   target commit. Pay particular attention to:
   - New files added by upstream (could introduce malicious code).
   - Changes to the NDJSON schema in `internal/output/` (breaks Beekeeper consumers).
   - Changes to the root resolver in `internal/resolver/` (conflicts with our Windows code paths).
   - Changes to `LICENSE` or `NOTICE` (legal compliance — we must preserve both verbatim or update
     our attribution accordingly).

3. **Run upstream tests on Linux and macOS.** Checkout the new upstream target commit in a separate
   worktree and run `go test ./...` on Linux and macOS to confirm upstream did not break itself
   before we absorb the change.

4. **Cherry-pick or merge preserving Windows code paths.** Cherry-pick or merge the new upstream
   commits into the Pollen branch, resolving conflicts in favor of preserving our Windows-specific
   files (`*_windows.go`, `cmd/pollen/`, `NOTICE`, `UPSTREAM.md`, `CHANGES.md`).

5. **Re-run the Pollen CI matrix.** Run `go test ./...` and `go vet ./...` on all three OSes
   (ubuntu-latest, macos-latest, windows-latest) to confirm the merged result is clean.

6. **Run the differential test.** The differential test asserts that `pollen` output on Linux is
   byte-for-byte identical to upstream `bumblebee` output for the same fixed fixture. Any divergence
   in the NDJSON output (excluding Windows-only fields) is a bug to fix before completing the merge.

7. **Bump the pinned commit and update CHANGES.md.** Update this file's `pinned commit` field with
   the new 40-char SHA, and append a new entry to `CHANGES.md` summarising what was absorbed.

8. **Tag a new Pollen release.** Tag the merged commit as `v0.1.1-pollen.N` (incrementing N), push
   the tag, and let the release workflow produce the signed binary + SBOM.

This procedure is also the two-person-rule discipline from PRD Section 12.2, applied to upstream
absorption: the diff-review step (step 2) is the primary gate.

## Version history

| Pollen version | Pinned upstream commit | Date absorbed | Summary |
|----------------|----------------------|---------------|---------|
| v0.1.1-pollen.1 | c24089804ee66ece4bec6f14638cb98985389cdb | 2026-05-22 | Initial fork — module path rewrite, cmd/pollen rename, Windows build clean, attribution files |
