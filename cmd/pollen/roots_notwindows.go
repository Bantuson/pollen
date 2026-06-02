//go:build !windows

package main

import "github.com/bantuson/pollen/internal/scanner"

// windowsBaselinePackageRoots is a no-op stub for non-Windows builds.
// The real implementation lives in roots_windows.go (//go:build windows).
// These stubs allow the case "windows": blocks in roots.go to compile on
// all platforms; the case is unreachable at runtime on non-Windows.
func windowsBaselinePackageRoots() []scanner.Root { return nil }

// windowsSystemRoots is a no-op stub for non-Windows builds.
func windowsSystemRoots() []scanner.Root { return nil }
