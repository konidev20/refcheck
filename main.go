package main

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/konidev20/refcheck/internal/template"
	"github.com/konidev20/refcheck/internal/ui"
	"github.com/konidev20/refcheck/internal/validator"
	"github.com/spf13/cobra"
)

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
	for _, t := range opts.Template {
		excludePatterns = append(excludePatterns, template.Templates[t].Exclude...)
	}
	combinedPattern := "(" + strings.Join(excludePatterns, ")|(") + ")"
	return regexp.MustCompile(combinedPattern)
}

func runChecker(cmd *cobra.Command, opts RefCheckOptions, _ []string) {
	folderPaths := opts.Paths
	numWorkers := opts.Workers
	jsonOutput := opts.JSON

	exclude := collectExcludePatterns(opts)

	results := make([]*validator.Result, len(folderPaths))

	for idx, folderPath := range folderPaths {
		result, err := validator.ProcessFolder(folderPath, exclude, numWorkers)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		results[idx] = result
	}

	ui.PrintResult(results, jsonOutput, cmd.OutOrStdout())
}
