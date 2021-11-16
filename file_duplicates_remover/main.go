package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = logger.Sync() }()

	err = readingFiles(*dirPath, logger)
	if err != nil {
		// Error has been logged in function readingFiles
		// No need to log here
		fmt.Println("Can't read directory. App will close in 3 seconds.")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}

	for i := 0; i < len(allFiles); i++ {
		checkFiles(i, logger)
	}
}

// readingFiles reads all files in directory given and in its subdirectories
func readingFiles(directoryPath string, logg *zap.Logger) error {
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		logg.Debug(
			fmt.Sprintf("Can't read directory '%s'", directoryPath),
			zap.Field{Key: "error", String: err.Error(), Type: zapcore.StringType},
		)

		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = readingFiles(strings.Join([]string{directoryPath, file.Name()}, "\\"), logg)
			if err != nil {
				// Error has been logged deeper in function readingFiles
				// No need to log here
				return err
			}
		} else {
			var f filesStruct
			fInfo, err := file.Info()
			if err != nil {
				logg.Debug(
					fmt.Sprintf("Error on reading file  '%s'", strings.Join([]string{directoryPath, file.Name()}, "\\")),
					zap.Field{Key: "error", String: err.Error(), Type: zapcore.StringType},
				)
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
// if delDuplicates flag is true, function asks which files to delete
func checkFiles(num int, logg *zap.Logger) {

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
				_, err := fmt.Scanln(&countDelete)
				if err != nil || countDelete > len(copiesNumber) {
					fmt.Println("Wrong count. Files not deleted")
					logg.Warn(
						fmt.Sprintf("Wrong count entered for  '%s' files. Files not deleted", allFiles[num].fileEntry.Name()),
					)

					return
				}
			}

			for k := 0; k < countDelete; k++ {
				fmt.Println("Enter number of file to delete. Enter 0 to save all files.")
				_, err := fmt.Scanln(&numberDelete)
				if err != nil || countDelete < numberDelete {
					fmt.Println("Wrong number. Files not deleted")
					logg.Warn(
						fmt.Sprintf("Wrong number entered for  '%s' files. Files not deleted", allFiles[num].fileEntry.Name()),
					)

					return
				}
				if numberDelete == 0 {
					return
				}
				os.Chdir(allFiles[copiesNumber[numberDelete-2]].filePath)
				err = os.Remove(allFiles[copiesNumber[numberDelete-2]].fileEntry.Name())
				if err != nil {
					fmt.Println("File not deleted. Error occured.")
					logg.Debug(
						fmt.Sprintf("Error on deleting file  '%s'", strings.Join([]string{allFiles[copiesNumber[numberDelete-2]].filePath, allFiles[copiesNumber[numberDelete-2]].fileEntry.Name()}, "\\")),
						zap.Field{Key: "error", String: err.Error(), Type: zapcore.StringType},
					)

				} else {
					fmt.Println("File deleted.")
					logg.Info(
						fmt.Sprintf("File deleted  '%s'", strings.Join([]string{allFiles[copiesNumber[numberDelete-2]].filePath, allFiles[copiesNumber[numberDelete-2]].fileEntry.Name()}, "\\")),
					)

				}
			}
		}
	}
}
