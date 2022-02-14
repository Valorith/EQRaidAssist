package scanner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Valorith/EQRaidAssist/loadFile"
	"github.com/Valorith/EQRaidAssist/player"
)

var (
	mu                sync.RWMutex
	isStarted         bool
	loadedRaidFile    string
	raidFrequency     time.Duration
	loadedLogFile     string
	logFrequency      time.Duration
	raidFrequencyChan chan int
	players           []*player.Player
	stopSignalChan    chan bool
	serverName        string
	characterName     string
)

func SetServerName(name string) error {
	mu.Lock()
	defer mu.Unlock()
	serverName = name
	return nil
}

func IsServerNameSet() bool {
	mu.RLock()
	defer mu.RUnlock()
	return len(serverName) > 0
}

func SetCharacterName(name string) error {
	mu.Lock()
	defer mu.Unlock()
	characterName = name
	return nil
}

func IsCharacterNameSet() bool {
	mu.RLock()
	defer mu.RUnlock()
	return len(characterName) > 0
}

// Starts the scanner, runs until told to stop
func Start() error {
	var err error
	mu.Lock()
	defer mu.Unlock()
	if isStarted {
		return nil
	}
	isStarted = true

	raidFrequency = 15 * time.Second
	logFrequency = 1 * time.Second
	raidFrequencyChan = make(chan int)
	stopSignalChan = make(chan bool)

	//raidScanner = FileScanner{loadedRaidFile, 1, false, 15} // Raid Dump Scan (Scan Type = 1)
	//logScanner = FileScanner{loadedLogFile, 2, false, 15}   // Log Scan (Scan Type = 2)

	//Load Players in

	err = scanRaid()
	if err != nil {
		fmt.Println("scanRaid failed:", err)
	}

	err = scanLog()
	if err != nil {
		fmt.Println("scanLog failed:", err)
	}

	go loop()
	return nil
}

// Stop stops the scanner
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	isStarted = false
	stopSignalChan <- true
}

// loops for as long as scanner is running (noted by isStarted)
func loop() {
	mu.RLock()
	raidTicker := time.NewTicker(raidFrequency)
	logTicker := time.NewTicker(logFrequency)
	mu.RUnlock()

	for {
		if !isStarted {
			return
		}
		select {
		case <-stopSignalChan:
			return
		case value := <-raidFrequencyChan:
			raidTicker = time.NewTicker(time.Duration(value) * time.Second)
		case <-raidTicker.C:
			err := scanRaid()
			if err != nil {
				fmt.Println("scanRaid failed:", err)
				continue
			}
		case <-logTicker.C:
			err := scanLog()
			if err != nil {
				fmt.Println("scanLog failed:", err)
				continue
			}
		}
	}
}

func scanRaid() error {
	if !isStarted {
		return fmt.Errorf("not started")
	}

	fmt.Println("ScanRaid started")

	newFileLocation, err := getNewestRaidFile()
	if err != nil {
		return fmt.Errorf("getNewestRaidFile: %w", err)
	}

	if loadedRaidFile == newFileLocation {
		return nil
	}
	loadedRaidFile = newFileLocation
	fmt.Println("Newest Raid Dump File Detected: ", loadedRaidFile)
	dumpLines, err := loadFile.Load(loadedRaidFile)
	if err != nil {
		return fmt.Errorf("loadFile.Load %s: %w", loadedRaidFile, err)
	}

	for _, line := range dumpLines {
		p, err := player.NewFromLine(line)
		if err != nil {
			return fmt.Errorf("player.NewFromLine failed (%s): %w", line, err)
		}
		players = append(players, p)
	}

	fmt.Printf("%+v\n", players)
	return nil
}

func scanLog() error {
	fmt.Println("log ran!")
	//logsFolder := EQpath + "\\Logs\\eqlog_" + charName + "_" + serverShortName + ".txt"
	return nil
}

func getNewestRaidFile() (string, error) {
	EQpath, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("getwd: %w", err)
	}
	//logsFolder := EQpath + "\\Logs"
	//fmt.Println("Loading Players from: ", EQpath)
	raidDumpFileList := []string{}
	fileListIndex := 0
	filePathError := filepath.Walk(EQpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("filepath.Walk: %w", err)
		}
		//fmt.Printf("Scanning File: %s....\n", path)
		isDir := info.IsDir()
		itemPath := path
		if !isDir {
			if strings.Contains(itemPath, "RaidRoster") {
				raidDumpFileList = append(raidDumpFileList, itemPath)
				fileListIndex++
			}
		}
		return nil
	})

	if filePathError != nil {
		fmt.Println(filePathError)
		return "", fmt.Errorf("filepath.Walk: %w", filePathError)
	}

	newestIndex := len(raidDumpFileList) - 1
	if newestIndex < 0 {
		return "", fmt.Errorf("empty result")
	}
	return raidDumpFileList[newestIndex], nil
}

func SetRaidFrequency(value int) {
	raidFrequencyChan <- value
}
