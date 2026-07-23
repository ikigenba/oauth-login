package main

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionFlagPrintsUnstampedVersionWithoutSideEffects(t *testing.T) {
	// R-9CGI-VDPQ
	var stdout, stderr bytes.Buffer
	launcher := &fakeLauncher{}

	if got := run(context.Background(), []string{"-V"}, &stdout, &stderr, testDependencies(launcher)); got != 0 {
		t.Fatalf("version exit = %d, want 0; stderr=%q", got, stderr.String())
	}
	if stdout.String() != "dev\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "dev\n")
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if launcher.count() != 0 {
		t.Fatalf("browser launches = %d, want 0", launcher.count())
	}
}

func TestStampedBinaryVersionFlagPrintsInjectedVersion(t *testing.T) {
	// R-9DOF-95GF
	const sentinel = "phase-07-stamped-version"
	binary := filepath.Join(t.TempDir(), "oauth-login")
	build := exec.Command("go", "build", "-ldflags", "-X main.version="+sentinel, "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build stamped oauth-login: %v\n%s", err, output)
	}

	command := exec.Command(binary, "-V")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("run stamped oauth-login: %v; stderr=%q", err, stderr.String())
	}
	if stdout.String() != sentinel+"\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), sentinel+"\n")
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestHelpNamesVersionFlagAndDescription(t *testing.T) {
	// R-9EWB-MX74
	var stdout, stderr bytes.Buffer
	if got := run(context.Background(), []string{"--help"}, &stdout, &stderr, testDependencies(&fakeLauncher{})); got != 0 {
		t.Fatalf("help exit = %d, want 0", got)
	}
	if stdout.Len() != 0 {
		t.Fatalf("help wrote stdout: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "  -V\n") || !strings.Contains(stderr.String(), "print version and exit") {
		t.Fatalf("help does not name -V with its description:\n%s", stderr.String())
	}
}
