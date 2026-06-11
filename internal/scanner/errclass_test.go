package scanner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"
	"testing"
)

// TestIsExpectedAccessError covers the routine "off-limits subtree" error
// classifier: permission sentinels and EACCES/EPERM (incl. wrapped) are
// expected; nil, ENOENT, and arbitrary errors are not.
func TestIsExpectedAccessError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"os.ErrPermission", os.ErrPermission, true},
		{"fs.ErrPermission", fs.ErrPermission, true},
		{"EACCES", syscall.EACCES, true},
		{"EPERM", syscall.EPERM, true},
		{"wrapped EACCES", fmt.Errorf("walk %s: %w", "dir", syscall.EACCES), true},
		{"ENOENT is not access", syscall.ENOENT, false},
		{"arbitrary", errors.New("boom"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isExpectedAccessError(tc.err); got != tc.want {
				t.Errorf("isExpectedAccessError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// TestIsMissingPathError covers the ENOENT / not-exist classifier used to
// downgrade routine per-host absent roots from warn to info.
func TestIsMissingPathError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"os.ErrNotExist", os.ErrNotExist, true},
		{"fs.ErrNotExist", fs.ErrNotExist, true},
		{"ENOENT", syscall.ENOENT, true},
		{"wrapped ENOENT", fmt.Errorf("stat %s: %w", "x", syscall.ENOENT), true},
		{"EACCES is not missing", syscall.EACCES, false},
		{"arbitrary", errors.New("boom"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isMissingPathError(tc.err); got != tc.want {
				t.Errorf("isMissingPathError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
