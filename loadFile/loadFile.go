package loadFile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

func Load(fileName string) ([]string, error) {
	var fileLines []string
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer file.Close()

	// Create a new Scanner for the file
	scanner := bufio.NewScanner(file)

	// Loop over all lines in the file and print them
	for scanner.Scan() {
		fileLines = append(fileLines, scanner.Text())
	}

	// Check for errors during Scan
	err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("scanner err: %w", err)
	}

	return fileLines, nil
}

// Get the size of the file at the provided path
func GetFileSize(fileName string) (int64, error) {
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

// Get the date/time when the file was last written to
func GetFileLastWrite(fileName string) (time.Time, error) {
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		return time.Time{}, err
	}

	return fileInfo.ModTime(), nil

}

// Moves the file from the sourcePath to the destPath
func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("MoveFile(): couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("MoveFile(): couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("MoveFile(): writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("MoveFile(): failed removing original file: %s", err)
	}
	return nil
}

// Returns true if the file at the provided filePath exists
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
