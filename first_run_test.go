package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsFirstRunForBuild(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "first_run_done")

	firstRun, err := isFirstRunForBuild(marker, "build-1")
	if err != nil {
		t.Fatalf("first build check failed: %v", err)
	}
	if !firstRun {
		t.Fatal("expected a new build to show the welcome page")
	}

	firstRun, err = isFirstRunForBuild(marker, "build-1")
	if err != nil {
		t.Fatalf("repeat build check failed: %v", err)
	}
	if firstRun {
		t.Fatal("expected the same build to skip the welcome page")
	}

	firstRun, err = isFirstRunForBuild(marker, "build-2")
	if err != nil {
		t.Fatalf("updated build check failed: %v", err)
	}
	if !firstRun {
		t.Fatal("expected a different build to show the welcome page")
	}

	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("could not read build marker: %v", err)
	}
	if string(data) != "build-2" {
		t.Fatalf("marker = %q, want %q", string(data), "build-2")
	}
}
