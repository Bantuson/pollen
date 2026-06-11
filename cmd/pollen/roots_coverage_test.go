package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bantuson/pollen/internal/model"
)

// TestClassifyRootRemainingArms covers the classifyRoot switch arms not already
// exercised by main_test.go (browser-extension, Firefox Profiles, homebrew, and
// the three profile fallbacks). Pure — path strings only.
func TestClassifyRootRemainingArms(t *testing.T) {
	cases := []struct {
		path, profile, want string
	}{
		{"/Users/a/Library/Application Support/Google/Chrome/Default/Extensions", model.ProfileBaseline, model.RootKindBrowserExtension},
		{"/Users/a/Library/Application Support/Firefox/Profiles", model.ProfileBaseline, model.RootKindBrowserExtension},
		{"/opt/homebrew/lib", model.ProfileBaseline, model.RootKindHomebrew},
		{"/usr/local/lib", model.ProfileBaseline, model.RootKindHomebrew},
		{"/Users/a/Library/Python", model.ProfileBaseline, model.RootKindHomebrew},
		{"/srv/random/dir", model.ProfileBaseline, model.RootKindUserPackage},
		{"/srv/random/dir", model.ProfileProject, model.RootKindProject},
		{"/srv/random/dir", "weird-profile", model.RootKindUnknown},
	}
	for _, c := range cases {
		if got := classifyRoot(c.path, c.profile); got != c.want {
			t.Errorf("classifyRoot(%q, %q) = %q, want %q", c.path, c.profile, got, c.want)
		}
	}
}

// TestClassifyRootDeepHome covers the isBroadHomeRoot arm of classifyRoot by
// pointing the home dir at a temp dir (USERPROFILE on Windows, HOME elsewhere).
func TestClassifyRootDeepHome(t *testing.T) {
	home := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	} else {
		t.Setenv("HOME", home)
	}
	if got := classifyRoot(home, model.ProfileDeep); got != model.RootKindDeepHome {
		t.Errorf("classifyRoot(home, deep) = %q, want %q", got, model.RootKindDeepHome)
	}
}

func TestContainsAny(t *testing.T) {
	if !containsAny("path/.vscode/extensions", "nope", ".vscode") {
		t.Error("containsAny should find .vscode")
	}
	if containsAny("plain/path", "x", "y") {
		t.Error("containsAny should be false when no sub matches")
	}
}

// TestUsersDirHelpers covers usersDirOverride + usersDirEffective via the
// POLLEN_USERS_DIR test seam.
func TestUsersDirHelpers(t *testing.T) {
	t.Setenv("POLLEN_USERS_DIR", "")
	if got := usersDirOverride(); got != "" {
		t.Errorf("usersDirOverride() unset = %q, want empty", got)
	}
	if got := usersDirEffective(); got != "/Users" {
		t.Errorf("usersDirEffective() unset = %q, want /Users", got)
	}

	t.Setenv("POLLEN_USERS_DIR", "  /tmp/fake-users  ")
	if got := usersDirOverride(); got != "/tmp/fake-users" {
		t.Errorf("usersDirOverride() = %q, want trimmed /tmp/fake-users", got)
	}
	if got := usersDirEffective(); got != "/tmp/fake-users" {
		t.Errorf("usersDirEffective() = %q, want override", got)
	}
}

// TestAllUsersHomes builds a fake /Users tree and asserts only real user-home
// directories survive the filter.
func TestAllUsersHomes(t *testing.T) {
	usersDir := t.TempDir()
	for _, d := range []string{"alice", "bob", "Shared", ".localized"} {
		if err := os.MkdirAll(filepath.Join(usersDir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{".DS_Store", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(usersDir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	homes := allUsersHomes(usersDir)
	got := map[string]bool{}
	for _, h := range homes {
		got[filepath.Base(h)] = true
	}
	if !got["alice"] || !got["bob"] {
		t.Errorf("allUsersHomes missing alice/bob: %v", homes)
	}
	for _, bad := range []string{"Shared", ".localized", ".DS_Store", "notes.txt"} {
		if got[bad] {
			t.Errorf("allUsersHomes should have filtered %q: %v", bad, homes)
		}
	}

	// Missing dir -> nil (ReadDir error branch).
	if h := allUsersHomes(filepath.Join(usersDir, "does-not-exist")); h != nil {
		t.Errorf("allUsersHomes(missing) = %v, want nil", h)
	}
}

// TestAllUsersExpansionNote exercises the diagnostic note (and, through it,
// allUsersHomes + usersDirEffective) via the override.
func TestAllUsersExpansionNote(t *testing.T) {
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "alice"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("POLLEN_USERS_DIR", usersDir)
	note := allUsersExpansionNote()
	if !strings.Contains(note, "expansion") || !strings.Contains(note, "1 home") {
		t.Errorf("allUsersExpansionNote() = %q, want it to mention expansion + 1 home", note)
	}
}

// TestHomesForExpansion covers the current-home path (the macOS --all-users
// fanout is platform-exclusive and not reachable on Windows/Linux).
func TestHomesForExpansion(t *testing.T) {
	for _, allUsers := range []bool{false, true} {
		homes := homesForExpansion(rootsOpts{AllUsers: allUsers})
		if len(homes) == 0 {
			t.Errorf("homesForExpansion(AllUsers=%v) returned no homes", allUsers)
		}
	}
}
