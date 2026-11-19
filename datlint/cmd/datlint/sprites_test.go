package main

import (
	"path/filepath"
	"testing"
)

func TestSpriteChecksCoverRepoSprites(t *testing.T) {
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	checks := spriteChecks(res)
	types := map[string]bool{}
	for _, c := range checks {
		types[c.Type] = true
		if c.HasErrors() {
			t.Errorf("%s %s: %+v", c.Type, c.Filename, c.Issues)
		}
	}
	if !types["rig"] || !types["flora"] {
		t.Fatalf("want rig and flora checks, got %v", types)
	}
	if spriteChecks("") != nil {
		t.Fatal("no res dir should mean no sprite checks")
	}
}
