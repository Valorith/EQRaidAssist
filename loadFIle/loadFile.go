package loadFile

import (
	"bufio"
	"fmt"
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
