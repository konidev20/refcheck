package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// processFile checks if the file is valid and calculates the SHA256 hash of the file
func processFile(filePath string, result *Result) {
	expectedHash := filepath.Base(filePath)
	result.TotalFiles++
	if !isValidSha256(expectedHash) {
		result.InvalidFiles++
		result.InvalidFileList = append(result.InvalidFileList, filePath)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		fmt.Printf("Error calculating SHA256 hash for file %s: %v\n", filePath, err)
		return
	}

	actualHash := hex.EncodeToString(hash.Sum(nil))

	if expectedHash == actualHash {
		result.IntactFiles++
	} else {
		result.CorruptedFiles++
		result.CorruptedFileList = append(result.CorruptedFileList, CorruptedFile{FilePath: filePath, ExpectedHash: expectedHash, ActualHash: actualHash})
	}
}

func isValidSha256(hash string) bool {
	// Check if the hash is 64 characters long
	if len(hash) != 64 {
		return false
	}

	// Check if the hash contains only hexadecimal digits
	if !regexp.MustCompile(`^[a-f0-9]+$`).MatchString(hash) {
		return false
	}
	return true
}
