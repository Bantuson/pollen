// Package endpoint collects host identity used in every record.
package endpoint

import (
	"os"
	"os/user"
	"runtime"
	"strconv"

	"github.com/bantuson/pollen/internal/model"
)

// Current returns the host identity used in every emitted record.
//
// deviceID, when non-empty, is set on Endpoint.DeviceID verbatim.
// Callers are expected to have already trimmed it and decided how to
// handle empty / whitespace input from the configured env var.
func Current(deviceID string) model.Endpoint {
	ep := model.Endpoint{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		DeviceID: deviceID,
	}
	if h, err := os.Hostname(); err == nil {
		ep.Hostname = h
	}
	if u, err := user.Current(); err == nil {
		ep.Username = u.Username
		if runtime.GOOS != "windows" {
			// On Windows, u.Uid is a SID string (S-1-5-21-...), not a numeric UID.
			// WPATH-02 requires endpoint.uid to be empty on Windows; leave ep.UID as zero value "".
			ep.UID = u.Uid
		}
	} else if runtime.GOOS != "windows" {
		// On Windows, os.Getuid() returns -1. WPATH-02 requires empty uid; skip.
		ep.UID = strconv.Itoa(os.Getuid())
	}
	return ep
}
