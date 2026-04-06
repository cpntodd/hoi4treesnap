package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteDiagnosticLogCreatesDatedFile(t *testing.T) {
	t.Setenv("APPIMAGE", "")
	tempDir := t.TempDir()
	oldBinPath := binPath
	binPath = tempDir
	defer func() { binPath = oldBinPath }()

	now := time.Date(2026, time.April, 6, 15, 4, 5, 0, time.Local)
	path, err := writeDiagnosticLogAt(now, "error.log", "boom")
	if err != nil {
		t.Fatalf("writeDiagnosticLogAt: %v", err)
	}

	expectedDir := filepath.Join(tempDir, "logs", "06-04-2026")
	if filepath.Dir(path) != expectedDir {
		t.Fatalf("diagnostic dir = %q, want %q", filepath.Dir(path), expectedDir)
	}

	if filepath.Base(path) != "03-04-05-06-04-2026-error.log" {
		t.Fatalf("diagnostic file = %q", filepath.Base(path))
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read diagnostic log: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "boom") {
		t.Fatalf("diagnostic log missing content: %q", text)
	}
	if !strings.Contains(text, "[2026-04-06T15:04:05") {
		t.Fatalf("diagnostic log missing timestamp header: %q", text)
	}
}

func TestWriteDiagnosticLogAvoidsSameSecondCollisions(t *testing.T) {
	t.Setenv("APPIMAGE", "")
	tempDir := t.TempDir()
	oldBinPath := binPath
	binPath = tempDir
	defer func() { binPath = oldBinPath }()

	now := time.Date(2026, time.April, 6, 15, 4, 5, 0, time.Local)
	firstPath, err := writeDiagnosticLogAt(now, "error.log", "first")
	if err != nil {
		t.Fatalf("first writeDiagnosticLogAt: %v", err)
	}
	secondPath, err := writeDiagnosticLogAt(now, "error.log", "second")
	if err != nil {
		t.Fatalf("second writeDiagnosticLogAt: %v", err)
	}

	if firstPath == secondPath {
		t.Fatalf("expected unique paths, got %q", firstPath)
	}
	if filepath.Base(secondPath) != "03-04-05-06-04-2026-error-01.log" {
		t.Fatalf("collision file = %q", filepath.Base(secondPath))
	}

	content, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatalf("read first diagnostic log: %v", err)
	}
	if strings.Contains(string(content), "second") {
		t.Fatalf("first diagnostic log was overwritten: %q", string(content))
	}
}