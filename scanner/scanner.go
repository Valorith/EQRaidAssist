package scanner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	startTime         []int  // Time the scanner was started
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

	err = setStartTime()
	if err != nil {
		return fmt.Errorf("Start: setStartTime: %w", err)
	}
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

		if err != nil {
			return fmt.Errorf("parseDateTime: %w", err)
		}
		if strings.Contains(lineText, "has been awarded to") { // Filter out non loot statements
			// Ensure the line occured after the start time
			lineRecent, err := checkRecent(lineText)
			if err != nil {
				return fmt.Errorf("scanLog: lineIsRecent: %w", err)
			}
			if !lineRecent {
				continue
			}

			charName, itemName, lootType, err := parseLootLine(lineText)
			if err != nil {
				return fmt.Errorf("scanLog: parseLootLine: %w", err)
			}
			fmt.Printf("%s has received %s from %s\n", charName, itemName, lootType)
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

func setStartTime() error {
	currentYear, currentMonth, currenteDay := time.Now().Date()
	currentHour, currentMinute, currentSecond := time.Now().Clock()
	var currentMonthInt int
	// convert month to int
	switch currentMonth.String() {
	case "January":
		currentMonthInt = 1
	case "February":
		currentMonthInt = 2
	case "March":
		currentMonthInt = 3
	case "April":
		currentMonthInt = 4
	case "May":
		currentMonthInt = 5
	case "June":
		currentMonthInt = 6
	case "July":
		currentMonthInt = 7
	case "August":
		currentMonthInt = 8
	case "September":
		currentMonthInt = 9
	case "October":
		currentMonthInt = 10
	case "November":
		currentMonthInt = 11
	case "December":
		currentMonthInt = 12
	default:
		return fmt.Errorf("setStartTime: invalid month: %s", currentMonth)
	}
	startTime = []int{currentYear, currentMonthInt, currenteDay, currentHour, currentMinute, currentSecond}
	return nil
}

func checkRecent(line string) (bool, error) {
	if !(len(startTime) == 6) {
		return false, fmt.Errorf("checkRecent: startTime is not set. length=%d", len(startTime))
	}
	startYear, startMonth, startDay, startHour, startMinute, startSecond := startTime[0], startTime[1], startTime[2], startTime[3], startTime[4], startTime[5]
	lineYear, lineMonth, lineDay, lineHour, lineMinute, lineSecond, err := parseDateTime(line)
	if err != nil {
		return false, fmt.Errorf("lineIsRecent: parseDateTime: %w", err)
	}

	// Debug Printout
	/*
		_, err = fmt.Printf("Line Date/Time: %d/%d/%d %d:%d:%d\n", lineMonth, lineDay, lineYear, lineHour, lineMinute, lineSecond)
		if err != nil {
			return false, fmt.Errorf("checkRecent: fmt.Println: %w", err)
		}
		_, err = fmt.Printf("Start Date/Time: %d/%d/%d %d:%d:%d\n", startMonth, startDay, startYear, startHour, startMinute, startSecond)
		if err != nil {
			return false, fmt.Errorf("checkRecent: fmt.Println: %w", err)
		}
	*/

	//Check if startTime is after the lineTime
	if startYear > lineYear {
		return false, nil
	}
	if startYear == lineYear && startMonth > lineMonth {
		return false, nil
	}
	if startYear == lineYear && startMonth == lineMonth && startDay > lineDay {
		return false, nil
	}
	if startYear == lineYear && startMonth == lineMonth && startDay == lineDay && startHour > lineHour {
		return false, nil
	}
	if startYear == lineYear && startMonth == lineMonth && startDay == lineDay && startHour == lineHour && startMinute > lineMinute {
		return false, nil
	}
	if startYear == lineYear && startMonth == lineMonth && startDay == lineDay && startHour == lineHour && startMinute == lineMinute && startSecond > lineSecond {
		return false, nil
	}
	return true, nil
}

func parseDateTime(line string) (int, int, int, int, int, int, error) {
	dateElements := strings.Split(line, " ")
	monthString := dateElements[1]
	dayString := dateElements[2]
	timeString := dateElements[3]
	yearString := dateElements[4][:len(dateElements[4])-1]

	day, err := strconv.Atoi(dayString)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseDateTime: strconv.Atoi: %w", err)
	}
	year, err := strconv.Atoi(yearString)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseDateTime: strconv.Atoi: %w", err)
	}
	hour, err := strconv.Atoi(timeString[:strings.Index(timeString, ":")])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseDateTime: strconv.Atoi: %w", err)
	}
	minute, err := strconv.Atoi(timeString[strings.Index(timeString, ":")+1 : strings.LastIndex(timeString, ":")])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseDateTime: strconv.Atoi: %w", err)
	}
	second, err := strconv.Atoi(timeString[strings.LastIndex(timeString, ":")+1:])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseDateTime: strconv.Atoi: %w", err)
	}

	var month int
	//convert month abreviation to number
	switch monthString {
	case "Jan":
		month = 1
	case "Feb":
		month = 2
	case "Mar":
		month = 3
	case "Apr":
		month = 4
	case "May":
		month = 5
	case "Jun":
		month = 6
	case "Jul":
		month = 7
	case "Aug":
		month = 8
	case "Sep":
		month = 9
	case "Oct":
		month = 10
	case "Nov":
		month = 11
	case "Dec":
		month = 12
	default:
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseDateTime: unknown month: %s", monthString)
	}
	return year, month, day, hour, minute, second, nil
}

func parseLootLine(line string) (string, string, string, error) {
	line = line[strings.Index(line, "]")+2:]
	itemName, playerName := "", ""
	if strings.Contains(line, "has been awarded to") {
		// Get the item name
		itemName = strings.TrimSpace(strings.Split(line, "has been awarded to")[0])
		// Get the player name
		playerName = strings.TrimSpace(strings.Split(line, "has been awarded to")[1])
		playerName = playerName[:strings.Index(playerName, " ")]
	} else if strings.Contains(line, "has been looted by the Master") {
		// Get the item name
		itemName = line[:strings.Index(line, "has been looted by the Master")-1]
		// Get the player name
		playerName = "Master Looter"
	}
	// Get loot distribution type
	masterLooterTake := "looted by the Master Loot."
	masterLooterTake = strings.ToLower(masterLooterTake)
	assignedTo := "by the Master Looter."
	assignedTo = strings.ToLower(assignedTo)
	randomTo := "by random roll."
	randomTo = strings.ToLower(randomTo)
	lcAward := "by the Loot Council."
	lcAward = strings.ToLower(lcAward)
	lootType := ""
	line = strings.ToLower(line)

	if strings.Contains(line, masterLooterTake) { // Matches if the master looter took the loot
		lootType = "master looting."
	} else if strings.Contains(line, assignedTo) { // Matches if the loot was assigned by the master looter
		lootType = "loot master assignment."
	} else if strings.Contains(line, randomTo) { // Matches if the loot was assigned by random roll
		lootType = "random assignment."
	} else if strings.Contains(line, lcAward) { // Matches if the loot was assigned by the loot council
		lootType = "the Loot Council"
	} else {
		fmt.Printf("Line: %s, compared with %s\n", line, masterLooterTake)
	}

	return playerName, itemName, lootType, nil
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
