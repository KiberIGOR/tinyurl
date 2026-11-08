package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateReset(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "./cmd/reset", "-root", root)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generator: %v\n%s", err, out)
	}

	gen := filepath.Join(root, "internal", "resettest", "reset.gen.go")
	data, err := os.ReadFile(gen)
	if err != nil {
		t.Fatal(err)
	}
	src := string(data)
	if !strings.Contains(src, "func (s *ResetableStruct) Reset()") {
		t.Fatal("missing ResetableStruct.Reset")
	}
	if !strings.Contains(src, "s.S = s.S[:0]") {
		t.Fatal("expected slice [:0] reset")
	}
	if !strings.Contains(src, "clear(s.M)") {
		t.Fatal("expected clear(map)")
	}
	if !strings.Contains(src, "interface{ Reset() }") {
		t.Fatal("expected nested Reset via interface")
	}
}
