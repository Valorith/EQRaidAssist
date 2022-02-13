package loadFile

import (
	"bufio"
	"log"
	"os"
	"time"
)

func Load(fileName string) []string {
	var fileLines []string
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a new Scanner for the file
	scanner := bufio.NewScanner(file)

	// Loop over all lines in the file and print them
	for scanner.Scan() {
		fileLines = append(fileLines, scanner.Text())
	}

	// Check for errors during Scan
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil
	}

	return fileLines
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
