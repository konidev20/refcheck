package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"
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

type RefCheckOptions struct {
	Paths    []string
	Exclude  []string
	Workers  int
	JSON     bool
	Template []string
}

var refCheckOptions RefCheckOptions

func main() {
	var rootCmd = &cobra.Command{
		Use:   "refcheck",
		Short: "refcheck checks the integrity of files in a directory",
		Long: `refcheck is a tool for checking the integrity of files in a directory.
Assuming the file names are the SHA256 hash of the file, it calculates the SHA256 hash of each file and compares it with the file name.
If the file name matches the hash, the file is intact; otherwise, it is corrupted.
The tool can be used to check the integrity of files in a directory before deploying them to a server.`,
		Run: func(cmd *cobra.Command, args []string) {
			runChecker(cmd, refCheckOptions, args)
		},
	}

	goos := runtime.GOOS

	rootCmd.Flags().StringSliceVarP(&refCheckOptions.Paths, "path", "p", []string{"."}, "Path to the folder. Can be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&refCheckOptions.Exclude, "exclude", "e", []string{}, "Regular expression pattern for excluding files and folders. Can be specified multiple times.")
	rootCmd.Flags().IntVarP(&refCheckOptions.Workers, "workers", "w", 4, "Number of workers for parallel processing")
	rootCmd.Flags().BoolVarP(&refCheckOptions.JSON, "json", "j", false, "Print the results in JSON format")
	rootCmd.Flags().StringSliceVarP(&refCheckOptions.Template, "template", "t", []string{"restic", goos}, "Template to use for excluding files and folders. Can be specified multiple times.")

	rootCmd.Execute()
}

// collectExcludePatterns compiles a regular expression that matches any of the file or folder patterns
// specified in the RefCheckOptions. This includes both directly specified exclude patterns and those
// derived from named templates.
func collectExcludePatterns(opts RefCheckOptions) *regexp.Regexp {
	excludePatterns := opts.Exclude
	for _, template := range opts.Template {
		excludePatterns = append(excludePatterns, templates[template].Exclude...)
	}
	combinedPattern := "(" + strings.Join(excludePatterns, ")|(") + ")"
	return regexp.MustCompile(combinedPattern)
}

func runChecker(cmd *cobra.Command, opts RefCheckOptions, _ []string) {
	folderPaths := opts.Paths
	numWorkers := opts.Workers
	jsonOutput := opts.JSON

	exclude := collectExcludePatterns(opts)

	results := make([]*Result, len(folderPaths))

	for idx, folderPath := range folderPaths {
		result, err := processFolder(folderPath, exclude, numWorkers)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		results[idx] = result
	}

	printResult(results, jsonOutput, cmd.OutOrStdout())
}

func processFolder(folderPath string, exclude *regexp.Regexp, numWorkers int) (*Result, error) {
	result := &Result{FolderPath: folderPath}

	var wg sync.WaitGroup
	fileChan := make(chan string)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				if !exclude.MatchString(filePath) {
					processFile(filePath, result)
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

func printResult(results []*Result, jsonOutput bool, w io.Writer) {
	if jsonOutput {
		jsonData, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		for _, result := range results {
			fmt.Println("")
			fmt.Println("-------------------")
			fmt.Println("Folder Path:", result.FolderPath)
			fmt.Println("")
			tbl := table.New("Result", "Value")
			tbl.WithHeaderSeparatorRow('-')
			tbl.WithPadding(10)
			tbl.WithWriter(w)

			tbl.AddRow("Total Files", result.TotalFiles)
			tbl.AddRow("Intact Files", result.IntactFiles)
			tbl.AddRow("Corrupted Files", result.CorruptedFiles)
			tbl.AddRow("Invalid Files", result.InvalidFiles)
			tbl.Print()
			fmt.Println("")
			fmt.Println("\nCorrupted Files:")
			if len(result.CorruptedFileList) > 0 {
				tbl = table.New("File Path", "Actual Hash")
				tbl.WithWriter(w)
				tbl.WithHeaderSeparatorRow('_')
				tbl.WithPadding(10)
				for _, file := range result.CorruptedFileList {
					tbl.AddRow(file.FilePath, file.ActualHash)
				}
			} else {
				fmt.Println("None")
			}
			fmt.Println("")
			fmt.Println("\nInvalid File Names:")
			if len(result.InvalidFileList) > 0 {
				tbl = table.New("File Path")
				tbl.WithWriter(w)
				tbl.WithHeaderSeparatorRow('-')
				tbl.WithPadding(10)
				for _, file := range result.InvalidFileList {
					tbl.AddRow(file)
				}

				tbl.Print()
			} else {
				fmt.Println("None")
			}
			fmt.Println("")
			fmt.Println("-------------------")
			fmt.Println("")
		}
	}
}
