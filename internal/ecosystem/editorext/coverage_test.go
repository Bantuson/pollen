package editorext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bantuson/pollen/internal/model"
)

// writeExt creates <root>/.<client>/extensions/<extDir>/package.json with body
// and returns (pjPath, extRoot, extDir).
func writeExt(t *testing.T, clientRoot, extDirName, body string) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	extRoot := filepath.Join(dir, clientRoot, "extensions")
	extDir := filepath.Join(extRoot, extDirName)
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pjPath := filepath.Join(extDir, "package.json")
	if err := os.WriteFile(pjPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return pjPath, extRoot, extDir
}

// TestScanExtensionFallbackPublisherDotName covers the dir-name fallback when
// package.json lacks name/version: "publisher.name-version" -> publisher+name.
// Uses a .vscode root so hostFromExtRoot's default (vscode) branch is hit too.
func TestScanExtensionFallbackPublisherDotName(t *testing.T) {
	pjPath, extRoot, extDir := writeExt(t, ".vscode", "ms-python.python-2024.1.0", `{}`)

	var out []model.Record
	s := &Scanner{MaxFileSize: 1 << 20, Emit: func(r model.Record) { out = append(out, r) }}
	if err := s.ScanExtension(pjPath, extRoot, extDir, model.Record{}); err != nil {
		t.Fatalf("ScanExtension: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("want 1 record, got %d", len(out))
	}
	if out[0].PackageName != "ms-python.python" || out[0].Version != "2024.1.0" {
		t.Errorf("fallback parse wrong: name=%q version=%q", out[0].PackageName, out[0].Version)
	}
	if out[0].PackageManager != "vscode" {
		t.Errorf("host = %q, want vscode (default branch)", out[0].PackageManager)
	}
}

// TestScanExtensionFallbackBareName covers the no-publisher-dot fallback branch
// ("name-version" with no '.').
func TestScanExtensionFallbackBareName(t *testing.T) {
	pjPath, extRoot, extDir := writeExt(t, ".vscodium", "soloext-9.9.9", `{}`)

	var out []model.Record
	s := &Scanner{MaxFileSize: 1 << 20, Emit: func(r model.Record) { out = append(out, r) }}
	if err := s.ScanExtension(pjPath, extRoot, extDir, model.Record{}); err != nil {
		t.Fatalf("ScanExtension: %v", err)
	}
	if len(out) != 1 || out[0].PackageName != "soloext" || out[0].Version != "9.9.9" {
		t.Fatalf("bare-name fallback wrong: %+v", out)
	}
	if out[0].PackageManager != "vscodium" {
		t.Errorf("host = %q, want vscodium", out[0].PackageManager)
	}
}

// TestScanExtensionErrors covers the three error returns: read error, JSON
// parse error, and incomplete-after-fallback.
func TestScanExtensionErrors(t *testing.T) {
	s := &Scanner{MaxFileSize: 1 << 20, Emit: func(model.Record) {}}

	t.Run("read error (missing file)", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), ".vscode", "extensions", "x.y-1.0.0", "package.json")
		if err := s.ScanExtension(missing, filepath.Dir(filepath.Dir(missing)), filepath.Dir(missing), model.Record{}); err == nil {
			t.Error("expected error for missing package.json")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		pjPath, extRoot, extDir := writeExt(t, ".vscode", "a.b-1.0.0", `{not json`)
		if err := s.ScanExtension(pjPath, extRoot, extDir, model.Record{}); err == nil {
			t.Error("expected parse error for invalid JSON")
		}
	})

	t.Run("incomplete after fallback", func(t *testing.T) {
		// dir name has no '-' so the fallback cannot recover name/version.
		pjPath, extRoot, extDir := writeExt(t, ".vscode", "noversiondir", `{}`)
		if err := s.ScanExtension(pjPath, extRoot, extDir, model.Record{}); err == nil {
			t.Error("expected incomplete-extension error")
		}
	})
}

// TestHostFromExtRoot covers every switch branch.
func TestHostFromExtRoot(t *testing.T) {
	cases := map[string]string{
		"/home/u/.cursor/extensions":        "cursor",
		"/home/u/.cursor-server/extensions": "cursor",
		"/home/u/.windsurf/extensions":      "windsurf",
		"/home/u/.vscodium/extensions":      "vscodium",
		"/home/u/.vscode-oss/extensions":    "vscodium",
		"/home/u/.vscode/extensions":        "vscode",
	}
	// NOTE: a backslash path (C:\...\.windsurf\...) is intentionally NOT asserted
	// here. hostFromExtRoot relies on filepath.ToSlash, which only rewrites the
	// HOST OS separator — so a backslash input is normalized on Windows but left
	// as-is on Linux/macOS, making such a case OS-dependent rather than a portable
	// assertion. The forward-slash cases above already cover every switch branch.
	for root, want := range cases {
		if got := hostFromExtRoot(root); got != want {
			t.Errorf("hostFromExtRoot(%q) = %q, want %q", root, got, want)
		}
	}
}

// TestReadBoundedEdgeCases covers the open-error, not-regular-file, and
// oversize branches of readBounded.
func TestReadBoundedEdgeCases(t *testing.T) {
	t.Run("open error", func(t *testing.T) {
		s := &Scanner{MaxFileSize: 1 << 20}
		if _, err := s.readBounded(filepath.Join(t.TempDir(), "nope.json")); err == nil {
			t.Error("expected open error for missing file")
		}
	})

	t.Run("not a regular file (directory)", func(t *testing.T) {
		s := &Scanner{MaxFileSize: 1 << 20}
		dir := t.TempDir() // a directory, not a regular file
		if _, err := s.readBounded(dir); err == nil {
			t.Error("expected not-a-regular-file error for a directory")
		}
	})

	t.Run("oversize file", func(t *testing.T) {
		dir := t.TempDir()
		big := filepath.Join(dir, "big.json")
		if err := os.WriteFile(big, make([]byte, 64), 0o644); err != nil {
			t.Fatal(err)
		}
		var diagMsg string
		s := &Scanner{MaxFileSize: 10, Diag: func(_, _, msg string) { diagMsg = msg }}
		if _, err := s.readBounded(big); err == nil {
			t.Error("expected oversize error")
		}
		if diagMsg == "" {
			t.Error("expected a Diag warning for the oversize skip")
		}
	})
}
