package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCurrentVersion(t *testing.T) {
	orig := Version
	t.Cleanup(func() { Version = orig })

	Version = "  9.9.9-test  "
	if got := currentVersion(); got != "9.9.9-test" {
		t.Errorf("currentVersion() with Version set = %q, want 9.9.9-test", got)
	}

	// Empty Version -> build info or compiled-in default; never empty.
	Version = ""
	if got := currentVersion(); got == "" {
		t.Error("currentVersion() with empty Version returned empty")
	}
}

func TestVersionString(t *testing.T) {
	s := versionString()
	for _, want := range []string{"pollen ", "commit:", "built:", "go:"} {
		if !strings.Contains(s, want) {
			t.Errorf("versionString() missing %q:\n%s", want, s)
		}
	}
}

func TestBuildHTTPAuth(t *testing.T) {
	for _, mode := range []string{"", "none"} {
		a, err := buildHTTPAuth(mode, "", "")
		if err != nil || a.Mode != "none" {
			t.Errorf("buildHTTPAuth(%q) = %+v, %v; want mode=none, no error", mode, a, err)
		}
	}

	// bearer
	if _, err := buildHTTPAuth("bearer", "", ""); err == nil {
		t.Error("bearer without token-env should error")
	}
	t.Setenv("PT_TOKEN_EMPTY", "")
	if _, err := buildHTTPAuth("bearer", "PT_TOKEN_EMPTY", ""); err == nil {
		t.Error("bearer with empty env value should error")
	}
	t.Setenv("PT_TOKEN", "secret")
	if a, err := buildHTTPAuth("bearer", "PT_TOKEN", ""); err != nil || a.Mode != "bearer" || a.Token != "secret" {
		t.Errorf("bearer = %+v, %v", a, err)
	}

	// hmac-sha256
	if _, err := buildHTTPAuth("hmac-sha256", "", ""); err == nil {
		t.Error("hmac without key-env should error")
	}
	t.Setenv("PT_HMAC_EMPTY", "")
	if _, err := buildHTTPAuth("hmac-sha256", "", "PT_HMAC_EMPTY"); err == nil {
		t.Error("hmac with empty env value should error")
	}
	t.Setenv("PT_HMAC", "key123")
	if h, err := buildHTTPAuth("hmac-sha256", "", "PT_HMAC"); err != nil || h.Mode != "hmac-sha256" || string(h.HMACKey) != "key123" {
		t.Errorf("hmac = %+v, %v", h, err)
	}

	if _, err := buildHTTPAuth("weird", "", ""); err == nil {
		t.Error("unknown auth mode should error")
	}
}

func TestOpenSink(t *testing.T) {
	t.Run("stdout and empty default", func(t *testing.T) {
		for _, dest := range []string{"stdout", ""} {
			w, closeFn, err := openSink(dest, "", false, sinkHTTPOpts{})
			if err != nil || w == nil || closeFn == nil {
				t.Fatalf("openSink(%q): w-nil=%v closeFn-nil=%v err=%v", dest, w == nil, closeFn == nil, err)
			}
			if err := closeFn(); err != nil {
				t.Errorf("stdout close: %v", err)
			}
		}
	})

	t.Run("file", func(t *testing.T) {
		if _, _, err := openSink("file", "", false, sinkHTTPOpts{}); err == nil {
			t.Error("file without --output-file should error")
		}
		fp := filepath.Join(t.TempDir(), "out.ndjson")
		for _, appendMode := range []bool{false, true} {
			w, closeFn, err := openSink("file", fp, appendMode, sinkHTTPOpts{})
			if err != nil || w == nil {
				t.Fatalf("file sink (append=%v): %v", appendMode, err)
			}
			if err := closeFn(); err != nil {
				t.Errorf("file close: %v", err)
			}
		}
		if _, _, err := openSink("file", filepath.Join(t.TempDir(), "missing-dir", "x.ndjson"), false, sinkHTTPOpts{}); err == nil {
			t.Error("file in a non-existent directory should error")
		}
	})

	t.Run("http", func(t *testing.T) {
		// bad auth mode propagates as an error
		if _, _, err := openSink("http", "", false, sinkHTTPOpts{AuthMode: "weird"}); err == nil {
			t.Error("http with bad auth mode should error")
		}
		// valid http sink construction (no connection happens until write)
		w, closeFn, err := openSink("http", "", false, sinkHTTPOpts{
			URL: "http://127.0.0.1:9", AuthMode: "none", AllowHTTP: true,
			BatchSize: 16, Timeout: time.Second, UserAgent: "pollen-test",
		})
		if err != nil {
			t.Fatalf("valid http sink construction errored: %v", err)
		}
		if w == nil || closeFn == nil {
			t.Fatal("http sink returned nil writer/close")
		}
		if err := closeFn(); err != nil {
			t.Errorf("http close: %v", err)
		}
	})

	t.Run("unknown dest", func(t *testing.T) {
		if _, _, err := openSink("carrier-pigeon", "", false, sinkHTTPOpts{}); err == nil {
			t.Error("unknown dest should error")
		}
	})
}
