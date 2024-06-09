# verifydata

verifydata is a command-line tool written in Go that checks the integrity of files within a specified directory. It assumes that the file names are the SHA256 hashes of their contents. The tool calculates the SHA256 hash of each file and compares it with the file name to determine if the file is intact or corrupted.

## Features

- **Parallel Processing:** Utilizes multiple workers to process files concurrently, improving performance on large datasets.
- **Exclusion Patterns:** Supports regular expressions to exclude specific files or directories from the check.
- **Output Options:** Can output results in a human-readable table format or as JSON for further processing.

## Installation

To install verifydata, you need to have Go installed on your machine. Follow these steps:

1. Clone the repository:
```
git clone https://github.com/konidev20/verifydata.git
```
2. Navigate to the project directory:
```
cd verifydata
```
3. Build the binary:
```
go build -o verifydata
```
4. Move the binary to a location in your PATH:
```
mv verifydata /usr/local/bin/verifydata
```

## Flags

- `-p, --path`: Specify the path to the directory you want to check. Default is the current directory.
- `-e, --exclude`: Provide regular expression patterns to exclude specific files or directories. This can be specified multiple times for multiple patterns.
- `-w, --workers`: Set the number of worker goroutines for processing files. Default is 4.
- `-j, --json`: Output the results in JSON format. By default, the output is in a human-readable table format.

## Example 1 - Table Format

```
verifydata -p . -e config -w 8
```

```
Result           Value  
------           -----  
Total Files      12     
Intact Files     10     
Corrupted Files  2      

Corrupted Files:
File Path                                  Expected Hash  Actual Hash                                                       
---------                                  -------------  -----------                                                       
/Users/konidev20/test-repo/.DS_Store       .DS_Store      486bc93ef63b3dec45192db542a0261efa88183924f63d1c8b90f891aba4c0d8  
/Users/konidev20/test-repo/data/.DS_Store  .DS_Store      c83cba8c16ff2edd1f38f406653ca72cc8cc6a42b357c8c31b742c8b3a1c9f65
```

## Example 2 - JSON Format

```
verifydata -p . -e config -w 8 -j
```

Output:
```
{
    "folder_path": "/Users/konidev20/test-repo",
    "total_files": 12,
    "intact_files": 10,
    "corrupted_files": 2,
    "corrupted_file_list": [
        {
            "file_path": "/Users/konidev20/test-repo/.DS_Store",
            "expected_hash": ".DS_Store",
            "actual_hash": "486bc93ef63b3dec45192db542a0261efa88183924f63d1c8b90f891aba4c0d8"
        },
        {
            "file_path": "/Users/konidev20/test-repo/data/.DS_Store",
            "expected_hash": ".DS_Store",
            "actual_hash": "c83cba8c16ff2edd1f38f406653ca72cc8cc6a42b357c8c31b742c8b3a1c9f65"
        }
    ]
}
```

## Templates
You can use templates to do verifydata for specific repsotitories. Currently, templates use exclusion
lists which are common to a specific repository format. For example, in restic repositories all files
except the `config` file have a SHA256 hash of the file name.

```
verifydata -p . -t restic -w 8 -j
```

The above command will exclude the restic repository specifc exclusion list and check the files.
