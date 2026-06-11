// Package coveragegate enforces Pollen's test-presence contract.
//
// Pollen is a Windows-only support fork of perplexityai/bumblebee. The gate
// (implemented as a test) asserts that every production .go file compiled on
// Windows lives in a package that has at least one test, OR is listed in a
// reason-coded allowlist. Files that are build-tagged out on Windows
// (//go:build unix, !windows, linux, darwin) are macOS/Linux territory —
// upstream Bumblebee's domain — and must be acknowledged explicitly in the
// allowlist rather than silently counting as "untested".
//
// This mirrors the beekeeper VAL-01 internal/coveragegate philosophy:
// presence + a fail-closed, reason-coded allowlist, NOT a single-OS percentage
// threshold (coverage on a Windows-only fork cannot meaningfully be expressed
// as one number — non-Windows branches are uncoverable here by construction).
package coveragegate
