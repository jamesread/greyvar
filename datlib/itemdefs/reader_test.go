package itemdefs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadItemdefFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shard.yml")
	err := os.WriteFile(path, []byte(`id: shard
title: Crystal Shard
stackMax: 50
category: material
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	item, err := ReadItemdefFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "shard" || item.Title != "Crystal Shard" || item.StackMax != 50 {
		t.Fatalf("unexpected item: %#v", item)
	}
}
