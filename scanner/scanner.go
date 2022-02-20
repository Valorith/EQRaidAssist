package scanner

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Valorith/EQRaidAssist/config"
	"github.com/Valorith/EQRaidAssist/core"
	"github.com/Valorith/EQRaidAssist/discord"
	"github.com/Valorith/EQRaidAssist/loadFile"
	"github.com/Valorith/EQRaidAssist/player"
	"github.com/Valorith/EQRaidAssist/raid"
	"github.com/hpcloud/tail"
)

var (
	mu                sync.RWMutex
	isStarted         bool          // Flags the scanner as active or inactive
	loadedRaidFile    string        // Directory of the raid dump file
	raidFrequency     time.Duration // Frequency of the raid dump file scan
	loadedLogFile     string        // Directory of the player's log file
	raidFrequencyChan chan int
	stopSignalChan    chan bool
	serverName        string // Server short name for reference in the log file directory
	characterName     string // Character name for reference in the log file directory
	startTime         []int  // Time the scanner was started
)

// Returns true if the file scanner is currently active
func IsRunning() bool {
	return isStarted
}

func SetServerName(name string) error {
	mu.Lock()
	defer mu.Unlock()
	serverName = name
	return nil
}

func GetServerName() string {
	mu.Lock()
	defer mu.Unlock()
	return serverName
}

func GetRaidTimer() string {
	mu.Lock()
	defer mu.Unlock()
	return raidFrequency.String()
}

func GetCharacterName() string {
	mu.Lock()
	defer mu.Unlock()
	return characterName
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
func Start() {
	var err error
	if isStarted {
		fmt.Println("scanner.Start(): scanner is already started")
		return
	}

	// Organize Raid Dump Files into subfolder
	OrganizeRaidDumps()

	isStarted = true
	err = setStartTime()
	if err != nil {
		fmt.Printf("scanner.Start(): setStartTime: %s", err)
	}
	raidFrequency = 10 * time.Second
	raidFrequencyChan = make(chan int)
	stopSignalChan = make(chan bool)

	loadSettings() // Load settings from config file

	//Load Players in
	err = scanRaid()
	if err != nil {
		fmt.Println("scanRaid failed:", err)
	}

	go loop()    // Loop through the scanner loop to detect new raid dump files
	go scanLog() // Start scanning character log for loot data
	fmt.Println("Listening...")
}

// Stop stops the scanner
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	isStarted = false
	stopSignalChan <- true
	fmt.Println("Scanner Shutting Down...")
}

// loops for as long as scanner is running (noted by isStarted)
func loop() {
	mu.RLock()
	raidTicker := time.NewTicker(raidFrequency)
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
		}
	}
}

func scanRaid() error {
	if !isStarted {
		return fmt.Errorf("not started")
	}

	// Organize Raid Dump Files into subfolder
	OrganizeRaidDumps()

	newFileLocation, fileLastModified, err := getNewestRaidFile()
	if err != nil {
		return fmt.Errorf("scanRaid: getNewestRaidFile: %w", err)
	}

	//fmt.Printf("File last Modified: %s -> %s\n", newFileLocation, fileLastModified)

	// Check if file was already loaded
	if loadedRaidFile == newFileLocation {
		//fmt.Println("Already operating on the newest Raid Dump...")
		return nil
	}

	// Ensure the newest file was created after the scanner started
	fileRecent, err := checkFileRecent(fileLastModified)
	if err != nil {
		return fmt.Errorf("scanRaid: checkFileRecent: %w", err)
	}
	if !fileRecent {
		//fmt.Printf("scanRaid: %s is not recent enough to load...\n", newFileLocation)
		return nil
	}

	loadedRaidFile = newFileLocation
	fmt.Println("Newest Raid Dump File Detected: ", loadedRaidFile)
	dumpLines, err := loadFile.Load(loadedRaidFile)
	if err != nil {
		return fmt.Errorf("loadFile.Load %s: %w", loadedRaidFile, err)
	}

	//Clear active players cache
	core.ClearPlayers()

	for _, line := range dumpLines {
		p, err := player.NewFromLine(line)
		if err != nil {
			return fmt.Errorf("player.NewFromLine failed (%s): %w", line, err)
		}
		// Add the player to the players cache
		core.AddPlayer(p)
		// Add the player to the raid if needed
		if !raid.PlayerIsInRaid(*p) {
			raid.AddPlayersToRaid()
		}
	}
	// Display loaded character cache
	fmt.Printf("%+v\n", core.Players)

	//Ensure all players are added to the active raid
	if raid.Active { // Start a check-in
		raid.AddPlayersToRaid()
		err = raid.ActiveRaid.CheckIn()
		if err != nil {
			return fmt.Errorf("raid.ActiveRaid.CheckIn: %w", err)
		}
	} else { // Start the raid
		raid.Start()
		raid.AddPlayersToRaid()
	}
	return nil
}

// Scans the character log for loot data
func scanLog() {
	fmt.Println("Log Scanner Booting Up...")
	// Establish the log filepath
	logFilePath, err := getLogDirectory()
	if err != nil {
		fmt.Printf("getLogDirectory: %s", err)
		isStarted = false
		return
	}
	// Monitor the character log file for loot messagess
	t, err := tail.TailFile(logFilePath, tail.Config{Follow: true})
	if err != nil {
		fmt.Printf("tail.TailFile: %s", err)
	}

	for line := range t.Lines {
		if !isStarted {
			fmt.Printf("scanLog: t.lines: exited log scan due to scanner being disabled")
			return
		}
		lineText := line.Text

		if isLootStatement(lineText) { // Filter out non loot statements
			// Ensure the line occured after the start time
			lineRecent, err := checkRecent(lineText)
			if err != nil {
				fmt.Printf("scanLog: lineIsRecent: %s", err)
			}
			if !lineRecent {
				continue
			}

			charName, itemName, lootType, err := parseLootLine(lineText)
			if err != nil {
				fmt.Printf("scanLog: parseLootLine: %s", err)
			}

			if lootType != "the Loot Council" { // Filter out all non loot council assignments
				continue
			}

			lootMessage := charName + " has received " + itemName + " from " + lootType
			fmt.Println(lootMessage)

			// Send discord message via WebHook
			discord.SendMessage(lootMessage, 1)

			// Assign loot to specific player
			for _, player := range core.Players {
				if player.Name == charName {
					err = player.AddLoot(itemName)
					if err != nil {
						fmt.Printf("scanLog: player.AddLoot: %s", err)
					}
					break
				}
			}
		}
	}
}

func loadSettings() error {
	err := config.ReadConfig()

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func setStartTime() error {
	currentYear, currentMonth, currenteDay := time.Now().Date()
	currentHour, currentMinute, currentSecond := time.Now().Clock()
	currentMonthInt, err := convertMonthToInt(currentMonth.String())
	if err != nil {
		return fmt.Errorf("setStartTime: convertMonthToInt: %w", err)
	}
	startTime = []int{currentYear, currentMonthInt, currenteDay, currentHour, currentMinute, currentSecond}
	return nil
}

func convertMonthToInt(monthString string) (int, error) {
	var currentMonthInt int
	// convert month to int
	switch monthString {
	case "January":
		currentMonthInt = 1
	case "Jan":
		currentMonthInt = 1
	case "February":
		currentMonthInt = 2
	case "Feb":
		currentMonthInt = 2
	case "March":
		currentMonthInt = 3
	case "Mar":
		currentMonthInt = 3
	case "April":
		currentMonthInt = 4
	case "Apr":
		currentMonthInt = 4
	case "May":
		currentMonthInt = 5
	case "June":
		currentMonthInt = 6
	case "Jun":
		currentMonthInt = 6
	case "July":
		currentMonthInt = 7
	case "Jul":
		currentMonthInt = 7
	case "August":
		currentMonthInt = 8
	case "Aug":
		currentMonthInt = 8
	case "September":
		currentMonthInt = 9
	case "Sep":
		currentMonthInt = 9
	case "October":
		currentMonthInt = 10
	case "Oct":
		currentMonthInt = 10
	case "November":
		currentMonthInt = 11
	case "Nov":
		currentMonthInt = 11
	case "December":
		currentMonthInt = 12
	case "Dec":
		currentMonthInt = 12
	default:
		return 0, fmt.Errorf("setStartTime: invalid month: %s", monthString)
	}
	return currentMonthInt, nil
}

// Detirmines if the line was created after the scanner started
func checkRecent(line string) (bool, error) {
	if !(len(startTime) == 6) {
		return false, fmt.Errorf("checkRecent: startTime is not set. length=%d", len(startTime))
	}
	startYear, startMonth, startDay, startHour, startMinute, startSecond := startTime[0], startTime[1], startTime[2], startTime[3], startTime[4], startTime[5]
	lineYear, lineMonth, lineDay, lineHour, lineMinute, lineSecond, err := parseDateTime(line)
	if err != nil {
		return false, fmt.Errorf("lineIsRecent: parseDateTime: %w", err)
	}

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

// Detirmines if the file was created after the scanner started
func checkFileRecent(fileTime string) (bool, error) {

	startYear, startMonth, startDay, startHour, startMinute, startSecond := startTime[0], startTime[1], startTime[2], startTime[3], startTime[4], startTime[5]
	fileYear, fileMonth, fileDay, fileHour, fileMinute, fileSecond, err := parseFileDateTime(fileTime)
	if err != nil {
		return false, fmt.Errorf("lineIsRecent: parseDateTime: %w", err)
	}

	//Check if startTime is after the lineTime
	if startYear > fileYear {
		return false, nil
	}
	if startYear == fileYear && startMonth > fileMonth {
		return false, nil
	}
	if startYear == fileYear && startMonth == fileMonth && startDay > fileDay {
		return false, nil
	}
	if startYear == fileYear && startMonth == fileMonth && startDay == fileDay && startHour > fileHour {
		return false, nil
	}
	if startYear == fileYear && startMonth == fileMonth && startDay == fileDay && startHour == fileHour && startMinute > fileMinute {
		return false, nil
	}
	if startYear == fileYear && startMonth == fileMonth && startDay == fileDay && startHour == fileHour && startMinute == fileMinute && startSecond > fileSecond {
		return false, nil
	}
	return true, nil
}

// Returns whether fileTime (first argument) is newer than oldFileTime (second argument)
func fileIsNewer(oldFileTime string, fileTime string) (bool, error) {

	oldYear, oldMonth, oldDay, oldHour, oldMinute, oldSecond, err := parseFileDateTime(oldFileTime)
	if err != nil {
		return false, fmt.Errorf("fileIsNewer: parseDateTime1: %w", err)
	}
	fileYear, fileMonth, fileDay, fileHour, fileMinute, fileSecond, err := parseFileDateTime(fileTime)
	if err != nil {
		return false, fmt.Errorf("fileIsNewer: parseDateTime2: %w", err)
	}

	//Check if startTime is after the lineTime
	if oldYear > fileYear {
		return false, nil
	}
	if oldYear == fileYear && oldMonth > fileMonth {
		return false, nil
	}
	if oldYear == fileYear && oldMonth == fileMonth && oldDay > fileDay {
		return false, nil
	}
	if oldYear == fileYear && oldMonth == fileMonth && oldDay == fileDay && oldHour > fileHour {
		return false, nil
	}
	if oldYear == fileYear && oldMonth == fileMonth && oldDay == fileDay && oldHour == fileHour && oldMinute > fileMinute {
		return false, nil
	}
	if oldYear == fileYear && oldMonth == fileMonth && oldDay == fileDay && oldHour == fileHour && oldMinute == fileMinute && oldSecond > fileSecond {
		return false, nil
	}
	return true, nil
}

// Returns the deconstructed date and time values for the provided log dateTime string
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

	month, err := convertMonthToInt(monthString)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseDateTime: convertMonthToInt: %w", err)
	}
	return year, month, day, hour, minute, second, nil
}

// Returns the deconstructed date and time values for the provided file dateTime string
func parseFileDateTime(dateTime string) (int, int, int, int, int, int, error) {
	if dateTime == "" {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseFileDateTime: dateTime is empty")
	}
	dateString := dateTime[:strings.Index(dateTime, " ")]
	timeString := dateTime[strings.Index(dateTime, " ")+1 : strings.Index(dateTime, ".")]

	//fmt.Printf("DateTime: %s\n", dateTime)
	//fmt.Printf("dateString: %s\n", dateString)
	//fmt.Printf("timeString: %s\n", timeString)
	// Get Date
	yearString := dateString[:strings.Index(dateString, "-")]
	dateString = dateString[strings.Index(dateString, "-")+1:]
	monthString := dateString[:strings.Index(dateString, "-")]
	dateString = dateString[strings.Index(dateString, "-")+1:]
	dayString := dateString

	// Convert to int
	day, err := strconv.Atoi(dayString)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseDateTime: strconv.Atoi: %w", err)
	}
	month, err := strconv.Atoi(monthString)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("parseDateTime: convertMonthToInt: %w", err)
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

	return year, month, day, hour, minute, second, nil
}

func isLootStatement(line string) bool {
	if strings.Contains(line, "has been awarded to") || strings.Contains(line, "has been looted by the Master") {
		return true
	}
	return false
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
	// Get the directory of the current executable
	mu.Lock()
	defer mu.Unlock()
	EQpath, err := os.Getwd() // Get the current working directory (used as EQpath)
	if err != nil {
		return "", fmt.Errorf("scanLog: os.Getwd: %w", err)
	}
	loadedLogFile = EQpath + "\\Logs\\eqlog_" + characterName + "_" + serverName + ".txt"
	return loadedLogFile, nil
}

// Return the directory to the newest detected Raid Dump file
func getNewestRaidFile() (string, string, error) {
	// Get Root EQ Directory
	EQpath, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return "", "", fmt.Errorf("getwd: %w", err)
	}

	// Get the directory of the RaidLogs folder
	raidLogsFolder := EQpath + "\\RaidLogs"

	// Load a list of all Raid dump files
	raidDumpFileList, err := getRaidDumpFiles(raidLogsFolder)
	if err != nil {
		return "", "", fmt.Errorf("getNewestRaidFile(): %w", err)
	}

	// Detirmine which file in raidDumpFileList is the newest
	newestIndex := 0
	newestModified := ""
	newestPath := ""
	for index, raidFilePath := range raidDumpFileList {
		oldFileModDate, err := getFileModDate(raidDumpFileList[newestIndex])
		if err != nil {
			return "", "", fmt.Errorf("getFileModDate: %w", err)
		}
		newFileModDate, err := getFileModDate(raidFilePath)
		if err != nil {
			return "", "", fmt.Errorf("getFileModDate: %w", err)
		}
		fileNewer, err := fileIsNewer(oldFileModDate, newFileModDate)
		if err != nil {
			return "", "", fmt.Errorf("getNewestRaidFile(): fileIsNewer: %w", err)
		}
		if fileNewer {
			newestIndex = index
			newestModified = newFileModDate
			newestPath = raidLogsFolder + "\\" + raidDumpFileList[index]
		}
	}
	if newestIndex < 0 {
		return "", "", fmt.Errorf("empty result")
	}

	return newestPath, newestModified, nil
}

func getRaidDumpFiles(basePath string) ([]string, error) {

	// Step through files and look for Raid Dump files
	raidDumpFileList := []string{}
	raidDumpFileInfo, err := ioutil.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("getRaidDumpFiles: %w", err)
	}

	// Add files to the file list
	for _, file := range raidDumpFileInfo {
		if strings.Contains(file.Name(), "RaidRoster") {
			fileName := file.Name()
			//fmt.Printf("Found Raid Dump file: %s...\n", fileName)
			raidDumpFileList = append(raidDumpFileList, fileName)
		}
	}

	//fmt.Printf("Returning a fileList with %d files from inside (%s)...\n", len(raidDumpFileList), basePath)
	return raidDumpFileList, nil
}

func getFileModDate(raidFilePath string) (string, error) {
	// Get the directory of the current executable
	EQpath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getFileModDate(): os.getwd: %w", err)
	}
	// Get the directory of the RaidLogs folder
	raidLogsFolder := EQpath + "\\RaidLogs"

	fileStat, err := os.Stat(raidLogsFolder + "\\" + raidFilePath)
	if err != nil {
		return "", fmt.Errorf("getFileModDate(%s): os.Stat: %w", raidFilePath, err)
	}
	fileModified := fileStat.ModTime().String()
	return fileModified, nil
}

func SetRaidFrequency(value int) {
	raidFrequencyChan <- value
}

// Organize raid dump files into a RaidLogs subfolder
func OrganizeRaidDumps() error {
	//fmt.Println("Starting file organization...")
	// Get the directory of the current executable
	EQpath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("organizeRaidDumps(): os.getwd: %w", err)
	}

	// Ensure that the RaidLog folder exists
	raidLogsFolder := EQpath + "\\RaidLogs"
	raidLogsFolderExists, err := loadFile.FileExists(raidLogsFolder)
	if err != nil {
		return fmt.Errorf("organizeRaidDumps(): loadFile.FileExists: %w", err)
	}
	if !raidLogsFolderExists {
		os.Mkdir(raidLogsFolder, 0777)
		fmt.Printf("Raid Log folder does not exist. Creating: %s\n", raidLogsFolder)
	}

	// Ensure that the SavedRaids folder exists
	savedRaidsFolder := EQpath + "\\SavedRaids"
	savedRaidsFolderExists, err := loadFile.FileExists(savedRaidsFolder)
	if err != nil {
		return fmt.Errorf("organizeRaidDumps(): loadFile.FileExists: %w", err)
	}
	if !savedRaidsFolderExists {
		os.Mkdir(savedRaidsFolder, 0777)
		fmt.Printf("SavedRaids folder does not exist. Creating: %s\n", savedRaidsFolder)
	}

	// Get the list of raid dump files
	raidDumpFileList, err := getRaidDumpFiles(EQpath)
	if err != nil {
		return fmt.Errorf("organizeRaidDumps(): getRaidDumpFiles: %w", err)
	}
	//fmt.Printf("%d logs found that need to be moved...\n", len(raidDumpFileList))
	// Loop through the raid dump files
	for _, raidFilePath := range raidDumpFileList {
		// Get the file name from the path
		raidFileName := filepath.Base(raidFilePath)
		// Move the file to the RaidLogs folder
		newFilePath := raidLogsFolder + "\\" + raidFileName
		err := loadFile.MoveFile(raidFilePath, newFilePath)
		if err != nil {
			return fmt.Errorf("organizeRaidDumps(): copyFile: %w", err)
		}
		fmt.Printf("Moving file: %s ---> %s\n", raidFilePath, newFilePath)
	}
	return nil
}
