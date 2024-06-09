package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/konidev20/verifydata/internal/template"
	"github.com/konidev20/verifydata/internal/ui"
	"github.com/konidev20/verifydata/internal/validator"
	"github.com/spf13/cobra"
)

type VerifyDataOptions struct {
	Paths     []string
	PathsFile []string
	Exclude   []string
	Workers   int
	JSON      bool
	Template  []string
}

var verifyDataOptions VerifyDataOptions

func main() {
	var rootCmd = &cobra.Command{
		Use:   "verifydata",
		Short: "verifydata checks the integrity of files in a directory",
		Long: `verifydata is a tool for checking the integrity of files in a directory.
Assuming the file names are the SHA256 hash of the file, it calculates the SHA256 hash of each file and compares it with the file name.
If the file name matches the hash, the file is intact; otherwise, it is corrupted.
The tool can be used to check the integrity of files in a directory before deploying them to a server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecker(cmd, verifyDataOptions, args)
		},
	}

	goos := runtime.GOOS

	rootCmd.Flags().StringSliceVarP(&verifyDataOptions.Paths, "path", "p", []string{"."}, "Path to the folder. Can be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&verifyDataOptions.PathsFile, "paths-file", "pf", []string{}, "Path to a file containing a list of folder paths. Each path should be on a new line.")
	rootCmd.Flags().StringSliceVarP(&verifyDataOptions.Exclude, "exclude", "e", []string{}, "Regular expression pattern for excluding files and folders. Can be specified multiple times.")
	rootCmd.Flags().IntVarP(&verifyDataOptions.Workers, "workers", "w", 4, "Number of workers for parallel processing")
	rootCmd.Flags().BoolVarP(&verifyDataOptions.JSON, "json", "j", false, "Print the results in JSON format")
	rootCmd.Flags().StringSliceVarP(&verifyDataOptions.Template, "template", "t", []string{"restic", goos}, "Template to use for excluding files and folders. Can be specified multiple times.")

	rootCmd.Execute()
}

// collectExcludePatterns compiles a regular expression that matches any of the file or folder patterns
// specified in the verifydataOptions. This includes both directly specified exclude patterns and those
// derived from named templates.
func collectExcludePatterns(opts VerifyDataOptions) *regexp.Regexp {
	excludePatterns := opts.Exclude
	for _, t := range opts.Template {
		excludePatterns = append(excludePatterns, template.Templates[t].Exclude...)
	}
	combinedPattern := "(" + strings.Join(excludePatterns, ")|(") + ")"
	return regexp.MustCompile(combinedPattern)
}

func getFolderPaths(opts VerifyDataOptions) ([]string, error) {
	folderPaths := opts.Paths
	for _, pf := range opts.PathsFile {
		file, err := os.Open(pf)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			folderPaths = append(folderPaths, scanner.Text())
		}
	}
	return folderPaths, nil
}

func runChecker(cmd *cobra.Command, opts VerifyDataOptions, _ []string) error {
	numWorkers := opts.Workers
	jsonOutput := opts.JSON

	folderPaths, err := getFolderPaths(opts)
	if err != nil {
		fmt.Printf("Error getting folder paths: %v\n", err)
		return err
	}

	exclude := collectExcludePatterns(opts)

	results := make([]*validator.Result, len(folderPaths))

	for idx, folderPath := range folderPaths {
		result, err := validator.ProcessFolder(folderPath, exclude, numWorkers)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}
		results[idx] = result
	}

	ui.PrintResult(results, jsonOutput, cmd.OutOrStdout())
	return nil
}
