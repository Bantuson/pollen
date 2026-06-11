package endpoint

import (
	"runtime"
	"testing"
)

// TestCurrentDeviceIDAndFields covers the Windows-reachable surface of
// Current: OS/Arch/DeviceID are always set, the device id is passed through
// verbatim, and an empty device id yields an empty DeviceID. (The numeric-UID
// branches are non-Windows-only and are reason-coded in the coverage gate.)
func TestCurrentDeviceIDAndFields(t *testing.T) {
	ep := Current("device-abc")
	if ep.DeviceID != "device-abc" {
		t.Errorf("DeviceID = %q, want %q", ep.DeviceID, "device-abc")
	}
	if ep.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", ep.OS, runtime.GOOS)
	}
	if ep.Arch != runtime.GOARCH {
		t.Errorf("Arch = %q, want %q", ep.Arch, runtime.GOARCH)
	}

	if empty := Current(""); empty.DeviceID != "" {
		t.Errorf("Current(\"\").DeviceID = %q, want empty", empty.DeviceID)
	}

	// On Windows, UID must be empty (WPATH-02: u.Uid is a SID, suppressed).
	if runtime.GOOS == "windows" && ep.UID != "" {
		t.Errorf("Windows endpoint UID = %q, want empty (WPATH-02)", ep.UID)
	}
}
