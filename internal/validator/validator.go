package validator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

type Result struct {
	FolderPath        string          `json:"folder_path"`
	TotalFiles        int             `json:"total_files"`
	IntactFiles       int             `json:"intact_files"`
	CorruptedFiles    int             `json:"corrupted_files"`
	CorruptedFileList []CorruptedFile `json:"corrupted_file_list"`
	InvalidFiles      int             `json:"invalid_files"`
	InvalidFileList   []string        `json:"invalid_file_list"`
}

type CorruptedFile struct {
	FilePath   string `json:"file_path"`
	ActualHash string `json:"actual_hash"`
}

// ValidateFile checks if the file is valid and calculates the SHA256 hash of the file
func ValidateFile(filePath string, result *Result) {
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
		result.CorruptedFileList = append(result.CorruptedFileList, CorruptedFile{FilePath: filePath, ActualHash: actualHash})
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

func ProcessFolder(folderPath string, exclude *regexp.Regexp, numWorkers int) (*Result, error) {
	result := &Result{FolderPath: folderPath}

	var wg sync.WaitGroup
	fileChan := make(chan string)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				if !exclude.MatchString(filePath) {
					ValidateFile(filePath, result)
				}
			}
		}()
	}

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileChan <- path
		}
		return nil
	})

	close(fileChan)
	wg.Wait()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil, err
	}

	return result, nil
}
