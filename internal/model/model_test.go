package model

import "testing"

// TestStableIDGoldenValues pins record_id outputs for representative records
// so future refactors of stableID cannot silently change the dedup key.
func TestStableIDGoldenValues(t *testing.T) {
	directDep := true

	pkg := Record{
		Profile:             ProfileBaseline,
		Ecosystem:           EcosystemNPM,
		NormalizedName:      "left-pad",
		Version:             "1.3.0",
		RootKind:            RootKindGlobalPackage,
		SourceType:          "package.json",
		SourceFile:          "/path/package.json",
		DirectDependency:    &directDep,
		HasLifecycleScripts: false,
		Confidence:          "high",
	}
	const wantPkg = "package:b6b6024a551185890759593eb31189c7744783e3efa471b1285b68451a60dfc3"
	if got := pkg.StableID(); got != wantPkg {
		t.Errorf("Record.StableID() = %q, want %q", got, wantPkg)
	}

	finding := Finding{
		Profile:        ProfileBaseline,
		FindingType:    FindingTypePackageExposure,
		CatalogID:      "cat-001",
		Ecosystem:      EcosystemNPM,
		NormalizedName: "left-pad",
		Version:        "1.3.0",
		RootKind:       RootKindGlobalPackage,
		SourceType:     "package.json",
		SourceFile:     "/path/package.json",
		Confidence:     "high",
	}
	const wantFinding = "finding:68b4cb2ed9440e2aacc37f145dffdfdfab2a9d36e395e8818d9954bd42f82d68"
	if got := finding.StableID(); got != wantFinding {
		t.Errorf("Finding.StableID() = %q, want %q", got, wantFinding)
	}

	diag := Diagnostic{
		Level:   "warn",
		Path:    "/some/path",
		Message: "skipped file",
	}
	const wantDiag = "diagnostic:88df9c032dda50d857e9d8809e20379c6dc8ff3f58746d9b8c93d489e648b4e3"
	if got := diag.StableID(); got != wantDiag {
		t.Errorf("Diagnostic.StableID() = %q, want %q", got, wantDiag)
	}
}

// TestStableIDDeterministic asserts two equal records hash to the same id and
// that distinct identity tuples produce distinct ids.
func TestStableIDDeterministic(t *testing.T) {
	a := Record{Profile: ProfileBaseline, Ecosystem: EcosystemNPM, NormalizedName: "x", Version: "1"}
	b := Record{Profile: ProfileBaseline, Ecosystem: EcosystemNPM, NormalizedName: "x", Version: "1"}
	c := Record{Profile: ProfileBaseline, Ecosystem: EcosystemNPM, NormalizedName: "x", Version: "2"}

	if a.StableID() != b.StableID() {
		t.Fatalf("equal records produced different ids: %q vs %q", a.StableID(), b.StableID())
	}
	if a.StableID() == c.StableID() {
		t.Fatalf("records with different versions hashed to the same id: %q", a.StableID())
	}
}

// TestSupportedEcosystems asserts the public ecosystem list is exactly the
// eight v0.1 values in canonical order, that the accessor returns a defensive
// copy (mutating the result must not corrupt the package-level order), and
// that IsSupportedEcosystem agrees with the list for both members and
// non-members.
func TestSupportedEcosystems(t *testing.T) {
	want := []string{
		EcosystemNPM,
		EcosystemPyPI,
		EcosystemGo,
		EcosystemRubyGems,
		EcosystemPackagist,
		EcosystemMCP,
		EcosystemEditorExtension,
		EcosystemBrowserExtension,
	}

	got := SupportedEcosystems()
	if len(got) != len(want) {
		t.Fatalf("SupportedEcosystems() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SupportedEcosystems()[%d] = %q, want %q", i, got[i], want[i])
		}
		if !IsSupportedEcosystem(want[i]) {
			t.Errorf("IsSupportedEcosystem(%q) = false, want true", want[i])
		}
	}

	// Defensive-copy contract: mutating the returned slice must not change a
	// subsequent call's result.
	got[0] = "mutated"
	if again := SupportedEcosystems(); again[0] != EcosystemNPM {
		t.Errorf("SupportedEcosystems() not a defensive copy: [0] = %q after caller mutation", again[0])
	}

	for _, bad := range []string{"", "cargo", "nuget", "NPM"} {
		if IsSupportedEcosystem(bad) {
			t.Errorf("IsSupportedEcosystem(%q) = true, want false", bad)
		}
	}
}

// TestRecordDedupKeyEqualsStableID locks the documented contract that the
// dedup key for a package record is its canonical record_id.
func TestRecordDedupKeyEqualsStableID(t *testing.T) {
	r := Record{
		Profile:        ProfileBaseline,
		Ecosystem:      EcosystemNPM,
		NormalizedName: "left-pad",
		Version:        "1.3.0",
		SourceFile:     "/path/package.json",
	}
	if r.DedupKey() != r.StableID() {
		t.Errorf("DedupKey() = %q, want StableID() = %q", r.DedupKey(), r.StableID())
	}
}

// TestRecordStableIDLifecycleScriptOrderInvariant proves StableID folds the
// lifecycle-script list through joinSorted: the same scripts in a different
// input order must yield the same id (exercises the sort branch of
// joinSorted, which the golden record left at 40%).
func TestRecordStableIDLifecycleScriptOrderInvariant(t *testing.T) {
	base := Record{
		Profile:             ProfileBaseline,
		Ecosystem:           EcosystemNPM,
		NormalizedName:      "build-tool",
		Version:             "2.0.0",
		HasLifecycleScripts: true,
	}
	a := base
	a.LifecycleScripts = []string{"postinstall", "install", "preinstall"}
	b := base
	b.LifecycleScripts = []string{"preinstall", "postinstall", "install"}

	if a.StableID() != b.StableID() {
		t.Errorf("lifecycle-script order changed StableID: %q vs %q", a.StableID(), b.StableID())
	}

	// A record with no scripts must differ from one with scripts.
	none := base
	none.HasLifecycleScripts = false
	if none.StableID() == a.StableID() {
		t.Errorf("record with no lifecycle scripts hashed equal to one with scripts: %q", none.StableID())
	}
}

// TestScanSummaryStableID exercises ScanSummary.StableID and, through it,
// canonicalCounts: equal summaries (including Counts built in different map
// insertion order, and Roots) hash equal, and a count change changes the id.
func TestScanSummaryStableID(t *testing.T) {
	mk := func(counts map[string]int) ScanSummary {
		return ScanSummary{
			Profile:               ProfileDeep,
			Status:                ScanStatusComplete,
			ScanTime:              "2026-06-11T00:00:00Z",
			EndTime:               "2026-06-11T00:01:00Z",
			Roots:                 []SummaryRoot{{Path: "/a", Kind: RootKindProject}, {Path: "/b", Kind: RootKindDeepHome}},
			Counts:                counts,
			PackageRecordsEmitted: 10,
			FindingsEmitted:       2,
			DurationMS:            1234,
		}
	}

	// Two logically-identical Counts maps built in different insertion order
	// must hash identically — this is what canonicalCounts' key sort buys.
	a := mk(map[string]int{"package": 10, "finding": 2})
	b := mk(map[string]int{"finding": 2, "package": 10})
	if a.StableID() != b.StableID() {
		t.Errorf("ScanSummary.StableID not invariant to Counts insertion order: %q vs %q", a.StableID(), b.StableID())
	}

	// A different count must change the id.
	c := mk(map[string]int{"package": 11, "finding": 2})
	if a.StableID() == c.StableID() {
		t.Errorf("ScanSummary.StableID ignored a count change: %q", a.StableID())
	}

	// Empty Counts must still produce a stable, scan_summary-prefixed id
	// (covers the len==0 branch of canonicalCounts).
	empty := mk(nil)
	if empty.StableID() != empty.StableID() {
		t.Error("ScanSummary.StableID not deterministic for empty Counts")
	}
	if got := empty.StableID(); got[:len(RecordTypeScanSummary)+1] != RecordTypeScanSummary+":" {
		t.Errorf("ScanSummary.StableID = %q, want %q-prefixed", got, RecordTypeScanSummary)
	}
}
