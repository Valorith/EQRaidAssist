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
	"github.com/hpcloud/tail"
)

var (
	mu                sync.RWMutex
	isStarted         bool          // Flags the scanner as active or inactive
	loadedRaidFile    string        // Directory of the raid dump file
	raidFrequency     time.Duration // Frequency of the raid dump file scan
	loadedLogFile     string        // Directory of the player's log file
	logFrequency      time.Duration // Frequency of the log file scan
	raidFrequencyChan chan int
	players           []*player.Player // Players detected within the raid dump file
	stopSignalChan    chan bool
	serverName        string // Server short name for reference in the log file directory
	characterName     string // Character name for reference in the log file directory
)

func SetServerName(name string) error {
	mu.Lock()
	defer mu.Unlock()
	serverName = name
	return nil
}

func GetServerName() string {
	return serverName
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

	newFileLocation, err := getNewestRaidFile()
	if err != nil {
		return fmt.Errorf("getNewestRaidFile: %w", err)
	}

	if loadedRaidFile == newFileLocation {
		fmt.Println("Already operating on the newest Raid Dump...")
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

// Scans the character log for loot data
func scanLog() error {
	// Establish the log filepath
	logFilePath, err := getLogDirectory()
	if err != nil {
		return fmt.Errorf("getLogDirectory: %w", err)
	}
	// Monitor the character log file for loot messages
	t, err := tail.TailFile(logFilePath, tail.Config{Follow: true})
	if err != nil {
		return fmt.Errorf("tail.TailFile: %w", err)
	}
	for line := range t.Lines {
		if !isStarted {
			return fmt.Errorf("scanLog: t.lines: exited log scan due to scanner being disabled")
		}
		lineText := line.Text
		if strings.Contains(lineText, "has been awarded to") {
			charName, itemName, err := parseLootLine(lineText)
			if err != nil {
				return fmt.Errorf("scanLog: parseLootLine: %w", err)
			}

			fmt.Printf("%s has received: %s\n", charName, itemName)
			for _, player := range players {
				if player.Name == charName {
					err = player.AddLoot(itemName)
					if err != nil {
						return fmt.Errorf("scanLog: player.AddLoot: %w", err)
					}
					break
				}
			}
		}
	}
	return nil
}

func parseLootLine(line string) (string, string, error) {
	// Get the item name
	//dateTime := line[:strings.Index(line, "]")+1]
	line = line[strings.Index(line, "]")+2:]
	itemName := strings.Split(line, "has been awarded to")[0]
	// Get the player name
	playerName := strings.TrimSpace(strings.Split(line, "has been awarded to")[1])
	playerName = playerName[:strings.Index(playerName, " ")]
	return playerName, itemName, nil
}

func getLogDirectory() (string, error) {
	EQpath, err := os.Getwd() // Get the current working directory (used as EQpath)
	if err != nil {
		return "", fmt.Errorf("scanLog: os.Getwd: %w", err)
	}
	loadedLogFile = EQpath + "\\Logs\\eqlog_" + characterName + "_" + serverName + ".txt"
	return loadedLogFile, nil
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
