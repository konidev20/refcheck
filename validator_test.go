package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidSha256(t *testing.T) {
	tests := []struct {
		name   string
		hash   string
		expect bool
	}{
		{"Valid Hash", "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc1", true},
		{"Invalid Hash Length", "abc123", false},
		{"Invalid Hash Characters", "xyz123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc1", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidSha256(test.hash); got != test.expect {
				t.Errorf("isValidSha256(%s) = %v, want %v", test.hash, got, test.expect)
			}
		})
	}
}

func TestProcessFile(t *testing.T) {
	// Setup

	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72")
	tempFile, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("test content")
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	defer tempFile.Close()

	result := &Result{}

	// Test valid file
	t.Run("Valid File", func(t *testing.T) {
		processFile(tempFile.Name(), result)
		if result.IntactFiles != 1 {
			t.Errorf("Expected 1 intact file, got %d", result.IntactFiles)
		}
	})

	result = &Result{}

	// Test invalid file name
	t.Run("Invalid File Name", func(t *testing.T) {
		processFile("invalidfilename", result)
		if result.InvalidFiles != 1 {
			t.Errorf("Expected 1 invalid file, got %d", result.InvalidFiles)
		}
	})

	result = &Result{}

	tempFile.WriteString("modifications")

	// Test corrupted file
	t.Run("Corrupted File", func(t *testing.T) {
		processFile(tempFile.Name(), result)
		if result.CorruptedFiles != 1 {
			t.Errorf("Expected 1 corrupted file, got %d", result.CorruptedFiles)
		}
	})
}
