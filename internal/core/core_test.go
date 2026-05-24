package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitAndJoin(t *testing.T) {
	dir := t.TempDir()
	originalPath := filepath.Join(dir, "testfile.bin")

	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	if err := os.WriteFile(originalPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	chunkSize := int64(256 * 1024)
	parts, err := Split(context.Background(), originalPath, chunkSize, nil)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	expectedParts := 4
	if len(parts) != expectedParts {
		t.Fatalf("expected %d parts, got %d", expectedParts, len(parts))
	}

	firstPart := parts[0]
	joinedPath, err := Join(context.Background(), firstPart, nil)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	joinedData, err := os.ReadFile(joinedPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(joinedData) != len(data) {
		t.Fatalf("joined size %d != original size %d", len(joinedData), len(data))
	}
	for i := range data {
		if joinedData[i] != data[i] {
			t.Fatalf("byte mismatch at offset %d: got %d, want %d", i, joinedData[i], data[i])
		}
	}
}

func TestSplitSmallFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "small.txt")
	content := []byte("hello world")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	parts, err := Split(context.Background(), path, 1024, nil)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}

	data, err := os.ReadFile(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Fatalf("unexpected content: %s", data)
	}
}
