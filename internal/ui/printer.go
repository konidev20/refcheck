package ui

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/konidev20/refcheck/internal/validator"
	"github.com/rodaine/table"
)

func PrintResult(results []*validator.Result, jsonOutput bool, w io.Writer) {
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
