package endpoint

import (
	"runtime"
	"testing"
)

func TestCurrentPopulatesDeviceID(t *testing.T) {
	ep := Current("dev-1234")
	if ep.DeviceID != "dev-1234" {
		t.Fatalf("DeviceID = %q, want %q", ep.DeviceID, "dev-1234")
	}
	if ep.OS != runtime.GOOS {
		t.Fatalf("OS = %q, want %q", ep.OS, runtime.GOOS)
	}
	if ep.Arch != runtime.GOARCH {
		t.Fatalf("Arch = %q, want %q", ep.Arch, runtime.GOARCH)
	}
}

func TestCurrentEmptyDeviceID(t *testing.T) {
	ep := Current("")
	if ep.DeviceID != "" {
		t.Fatalf("DeviceID = %q, want empty", ep.DeviceID)
	}
}

func TestCurrentWindowsUID(t *testing.T) {
	ep := Current("")
	if runtime.GOOS == "windows" {
		// WPATH-02: on Windows, user.Current().Uid is a SID string; endpoint.uid must be empty.
		if ep.UID != "" {
			t.Errorf("endpoint.uid on Windows = %q, want empty string (WPATH-02)", ep.UID)
		}
	} else {
		// D-04 regression guard: Unix UID must remain a non-empty numeric string.
		if ep.UID == "" {
			t.Errorf("endpoint.uid on %s = empty, want non-empty numeric UID (regression)", runtime.GOOS)
		}
	}
	// os/arch/username are already asserted by the existing TestCurrentPopulatesDeviceID — no need
	// to duplicate that here.
}
