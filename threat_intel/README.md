# threat_intel/

This directory ships **empty by design**.

## Why is it empty?

Beekeeper consumes upstream threat-intelligence catalogs directly via
`beekeeper catalogs sync`, which pulls from the upstream source over HTTPS as
part of normal operation (PRD §6.3 "reference" decision). Duplicating catalogs
here would create two paths for the same data with two possible drift conditions.
Single-source is cleaner.

## Compatibility

The `--exposure-catalog` CLI flag is preserved for compatibility with upstream
Bumblebee's interface. It accepts a path to a catalog JSON file or directory.
Pointing it at this directory returns no results (empty catalog) by default.

To use live threat intelligence catalogs, run `beekeeper catalogs sync` from
the [Beekeeper](github.com/bantuson/beekeeper) harness. Beekeeper manages the
catalog lifecycle and passes catalog data to Pollen scans automatically.

## Selftest fixtures

Upstream selftest fixtures (used by `pollen selftest`) live under
`cmd/pollen/selftest/` — not in this directory. Those fixtures are embedded at
build time and are not catalogs in the threat-intel sense.
