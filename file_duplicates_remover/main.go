package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	delDuplicates *bool
	dirPath       *string
	allFiles      []filesStruct
)

type filesStruct struct {
	fileEntry   os.DirEntry
	fileSize    int64
	filePath    string
	fileChecked bool
}

func init() {
	delDuplicates = flag.Bool("delDuplicates", false, "Delete duplicates? false=No  true=Yes")
	dirPath = flag.String("dirPath", "D:\\", "Path to directory for inspection. Program is configured to work on OS Windows.")
	flag.Parse()
}

func main() {
	err := readingFiles(*dirPath)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(allFiles); i++ {
		checkFiles(i)
	}
}

// readingFiles reads all files in directory given and in its subdirectories
func readingFiles(directoryPath string) error {
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = readingFiles(strings.Join([]string{directoryPath, file.Name()}, "\\"))
			if err != nil {
				return err
			}
		} else {
			var f filesStruct
			fInfo, err := file.Info()
			if err != nil {
				return err
			}
			f.fileEntry = file
			f.filePath = directoryPath
			f.fileChecked = false
			f.fileSize = fInfo.Size()
			allFiles = append(allFiles, f)
		}
	}
	return nil
}

// checkFiles checks if file on given position in slice have copies
// if delDuplicates flag is true, function ask which files to delete
func checkFiles(num int) {

	var copiesNumber []int
	foundCopy := false

	if allFiles[num].fileChecked {
		return
	}
	for j := num + 1; j < len(allFiles); j++ {
		if allFiles[num].fileEntry.Name() == allFiles[j].fileEntry.Name() && allFiles[num].fileSize == allFiles[j].fileSize {
			copiesNumber = append(copiesNumber, j)
			foundCopy = true
			allFiles[j].fileChecked = true
		}
	}
	if foundCopy {
		if allFiles[num].fileChecked {
			return
		}
		fmt.Println("Found copies: \n1.", allFiles[num].fileEntry.Name(), "    ", allFiles[num].filePath)
		for j := 0; j < len(copiesNumber); j++ {
			fmt.Print(j + 2)
			fmt.Println(". ", allFiles[copiesNumber[j]].fileEntry.Name(), "    ", allFiles[copiesNumber[j]].filePath)
		}
		if *delDuplicates {
			countDelete := 1
			var numberDelete int
			if len(copiesNumber) > 1 {
				fmt.Println("Enter count of files to delete. Enter 0 to save all files.")
				fmt.Scanln(&countDelete)
			}

			for k := 0; k < countDelete; k++ {
				fmt.Println("Enter number of file to delete. Enter 0 to save all files.")
				fmt.Scanln(&numberDelete)
				if numberDelete == 0 {
					return
				}
				os.Chdir(allFiles[copiesNumber[numberDelete-2]].filePath)
				err := os.Remove(allFiles[copiesNumber[numberDelete-2]].fileEntry.Name())
				if err != nil {
					fmt.Println("File not deleted. Error occured.")
					log.Println(err)
				} else {
					fmt.Println("File deleted.")
				}
			}
		}
	}
}
