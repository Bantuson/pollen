package rubygems

import "testing"

// TestIsGemfileLockAndGemspec pins the two filename predicates used to route
// candidate files to the RubyGems scanners.
func TestIsGemfileLockAndGemspec(t *testing.T) {
	lockCases := map[string]bool{
		"Gemfile.lock": true,
		"gemfile.lock": false, // case-sensitive
		"Gemfile":      false,
		"Gemfile.lock.bak": false,
		"":             false,
	}
	for base, want := range lockCases {
		if got := IsGemfileLock(base); got != want {
			t.Errorf("IsGemfileLock(%q) = %v, want %v", base, got, want)
		}
	}

	gemspecCases := map[string]bool{
		"foo.gemspec":      true,
		"a-b_c.gemspec":    true,
		"foo.gemspec.tmpl": false,
		"foo.gem":          false,
		"gemspec":          false,
		"":                 false,
	}
	for base, want := range gemspecCases {
		if got := IsGemspec(base); got != want {
			t.Errorf("IsGemspec(%q) = %v, want %v", base, got, want)
		}
	}
}
