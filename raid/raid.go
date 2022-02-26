package raid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Valorith/EQRaidAssist/alias"
	"github.com/Valorith/EQRaidAssist/core"
	"github.com/Valorith/EQRaidAssist/mongodb"
	"github.com/Valorith/EQRaidAssist/player"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mu          sync.Mutex
	Active      bool
	ActiveRaid  Raid
	AllRaids    RaidCollection = RaidCollection{}
	DisplayList map[string]int = map[string]int{}
)

func ResetData() {
	mu.Lock()
	defer mu.Unlock()
	Active = false
	ActiveRaid = Raid{}
}

type RaidCollection struct {
	RaidList []Raid `json:"raidList"`
}

type Raid struct {
	Name        string           `json:"name"`        // Name of the raid
	StartYear   int              `json:"startyear"`   // Start year of the raid
	StartMonth  int              `json:"startmonth"`  // Start month of the raid
	StartDay    int              `json:"startday"`    // Start day of the raid
	StartHour   int              `json:"starthour"`   // Start day of the raid
	StartMinute int              `json:"startminute"` // Start day of the raid
	StartSecond int              `json:"startsecond"` // Start day of the raid
	Description string           `json:"description"` // Raid description
	Checkins    map[string]int   `json:"checkins"`    // Map of raid check-ins for each respective member [player_anme]checkIns
	Players     []*player.Player `json:"players"`     // List of players in the raid
	FileName    string           `json:"filename"`
	Active      bool             `json:"active"` // Indicates whether the raid is active or not
}

func (raid *RaidCollection) AddRaid(newRaid Raid) error {
	if !Active {
		return fmt.Errorf("Raid is not active")
	}
	raid.RaidList = append(raid.RaidList, newRaid)
	return nil
}

func (raids *RaidCollection) UpdateDB() error {
	if !mongodb.RaidsDB.Connected {
		return fmt.Errorf("RaidCollection: UpdateDB(): raid database not connected")
	}
	fmt.Println("Dropping Raids Collection...")
	mongodb.RaidsDB.Collection.Drop(mongodb.RaidsDB.Context)
	for _, raid := range raids.RaidList {
		fmt.Printf("Updating DB with raid: %s\n", raid.Name)
		mongodb.RaidsDB.Insert(raid)
	}
	return nil
}

func (raids *RaidCollection) LoadFromDB() error {
	// Ensure the database is connected
	if !mongodb.RaidsDB.Connected {
		return fmt.Errorf("LoadFromDB(): alias database not connected")
	}

	// Get the list of raids from the database
	//opts := options.Find().SetSort(bson.D{}) // example: {Key: "handle", Value: Valgor}
	opts := options.Find().SetProjection(bson.D{{Key: "_id", Value: 0}}) // Ignore _id field
	loadedData, err := mongodb.RaidsDB.Collection.Find(mongodb.RaidsDB.Context, bson.M{}, opts)
	if err != nil {
		return fmt.Errorf("LoadFromDB(): mongodb.RaidsDB.Collection.Find(): %w", err)
	}

	// Get all entires from the returned data
	var loadedRaids []Raid
	if err = loadedData.All(mongodb.RaidsDB.Context, &loadedRaids); err != nil {
		return fmt.Errorf("LoadFromDB(): loadedData.All(): %w", err)
	}

	// Add all loaded raids to the raid collection
	raids.RaidList = loadedRaids

	return nil
}

func Start() error {
	if Active {
		return fmt.Errorf("Raid is already active")
	}
	Active = true
	currentYear, currentMonth, currenteDay := time.Now().Date()
	currentHour, currentMinute, currentSecond := time.Now().Clock()
	activePlayers := core.GetActivePlayers()
	// Initialize Active Raid struct
	ActiveRaid = Raid{
		Name:        "RaidAttend_" + strconv.Itoa(currentYear) + "-" + strconv.Itoa(int(currentMonth)) + "-" + strconv.Itoa(currenteDay) + "-" + strconv.Itoa(currentHour) + strconv.Itoa(currentMinute),
		StartYear:   currentYear,
		StartMonth:  int(currentMonth),
		StartDay:    currenteDay,
		StartHour:   currentHour,
		StartMinute: currentMinute,
		StartSecond: currentSecond,
		Description: "",
		Checkins:    make(map[string]int),
		Players:     activePlayers,
		FileName:    "RaidAttend_" + strconv.Itoa(currentYear) + "-" + strconv.Itoa(int(currentMonth)) + "-" + strconv.Itoa(currenteDay) + "-" + strconv.Itoa(currentHour) + strconv.Itoa(currentMinute) + ".json",
		Active:      true}
	//-----------------------
	ActiveRaid.initializeCheckins()
	ActiveRaid.Save()
	return nil
}

// Create a func that displays all entries in the DisplayList map
func PrintDisplayList() error {
	if len(DisplayList) > 0 {
		for character, checkins := range DisplayList {
			fmt.Printf("%s: %d\n", character, checkins)
		}
	} else {
		return fmt.Errorf("PrintDisplayList(): DisplayList is empty")
	}
	return nil
}

func (raid Raid) initializeCheckins() {
	for _, player := range raid.Players {
		ActiveRaid.Checkins[player.Name] = 1
	}
	//fmt.Printf("%d checkins initialized at a count of 1...\n", len(ActiveRaid.Checkins))
}

func (raid Raid) CheckIn() error {

	//Ensure Everyone in the raid is the checkins map (default to 0)
	for _, player := range raid.Players {
		if _, ok := raid.Checkins[player.Name]; !ok {
			raid.Checkins[player.Name] = 0
		}
	}

	//Increment checkinCounts
	for playerName, checkIns := range raid.Checkins {
		//Ensure player is on the current player list
		if playerIsInCache(playerName) {
			raid.Checkins[playerName] = checkIns + 1
		}
	}
	ActiveRaid.Save()
	return nil
}

func (raid Raid) PrintLoot() {
	// Display the loot for the raid
	for _, player := range raid.Players {
		fmt.Printf("Name: %s\nLoot:\n", player.Name)
		for index, lootItem := range player.Loot {
			fmt.Printf("%d) %s\n", index+1, lootItem.Name)
		}
	}
}

func (raid Raid) GetPlayerByName(playerName string) *player.Player {
	// Get player by name
	for _, player := range raid.Players {
		if player.Name == playerName {
			return player
		}
	}
	return nil
}

func (raid Raid) PrintParticipation() error {
	if !raid.Active {
		return fmt.Errorf("Raid is not active")
	}
	if len(raid.Players) == 0 {
		return fmt.Errorf("there are no players in the raid")
	}
	fileName := raid.Name + ".json"
	fmt.Println("Print Raid: " + fileName)
	//Increment checkinCounts
	fmt.Println("Checkin Count:", len(ActiveRaid.Checkins))
	index := 1
	for playerName, checkIns := range ActiveRaid.Checkins {
		//fmt.Printf("%s has checked in %d times\n", playerName, checkIns)
		//Ensure player is on the current player list
		if playerIsInCache(playerName) {
			fmt.Printf("%d) %s: %d [Active]\n", index, playerName, checkIns)
		} else {
			fmt.Printf("%d) %s: %d [Inactive]\n", index, playerName, checkIns)
		}
		index++
	}
	return nil
}

func (raid Raid) Save() error {
	err := SaveRaid(raid)
	if err != nil {
		return fmt.Errorf("error saving raid: %s", err)
	}
	return nil
}

func (raid Raid) GetCheckinsByName(name string) int {
	mu.Lock()
	defer mu.Unlock()
	return raid.Checkins[name]
}

func (raid Raid) Load(savedRaid string) error {
	var err error
	if savedRaid == "" {
		savedRaid, err = getLastRaidFile()
		if err != nil {
			return fmt.Errorf("getLastRaidFile(): error loading raid: %s", err)
		}
		if savedRaid == "" {
			return fmt.Errorf("Raid.Load(savedRaid): no raid files found or provided")
		}
	}

	loadedRaid, err := LoadRaid(savedRaid)
	if err != nil {
		return fmt.Errorf("error loading raid: %s", err)
	}
	ActiveRaid = loadedRaid
	core.Players = ActiveRaid.Players

	fmt.Println("Loaded Raid: " + ActiveRaid.Name)
	return nil
}

// Returns the directory to the most recent saved raid file
func getLastRaidFile() (string, error) {
	// Get a slice of all saved raid files
	savedRaidFiles, err := getSavedRaidFiles()
	if err != nil {
		return "", fmt.Errorf("getLastRaidFile(): error getting saved raid files: %s", err)
	}
	if len(savedRaidFiles) == 0 {
		return "", fmt.Errorf("getLastRaidFile(): no raid files found")
	}

	// Iterate through all saved raid files and detirmine the newest
	newestIndex := 0
	for index, fileName := range savedRaidFiles {
		modDate, err := getSavedRaidModDate(fileName)
		if err != nil {
			return "", fmt.Errorf("getLastRaidFile(): getSavedRaidModDate(%s): error getting saved raid mod date: %s", fileName, err)
		}
		newestModDate, err := getSavedRaidModDate(savedRaidFiles[newestIndex])
		if err != nil {
			return "", fmt.Errorf("getLastRaidFile(): getSavedRaidModDate(%s): error getting saved raid mod date: %s", savedRaidFiles[newestIndex], err)
		}
		fileNewer, err := fileIsNewer(modDate, newestModDate)
		if err != nil {
			return "", fmt.Errorf("getLastRaidFile(): fileIsNewer(%s, %s): error comparing saved raid mod dates: %s", modDate, savedRaidFiles[newestIndex], err)
		}
		if fileNewer {
			newestIndex = index
		}
	}

	// Return the newest saved raid file
	return savedRaidFiles[newestIndex], nil
}

func getSavedRaidFiles() ([]string, error) {
	// Get the directory of the current executable
	EQpath, err := os.Getwd()
	if err != nil {
		return []string{}, fmt.Errorf("getSavedRaidFiles(): os.getwd: %w", err)
	}
	// Get the directory of the RaidLogs folder
	savedRaidsFolder := EQpath + "\\SavedRaids"

	// Step through files and look for Raid Dump files
	savedRaidsFileList := []string{}
	savedRaidsFileInfo, err := ioutil.ReadDir(savedRaidsFolder)
	if err != nil {
		return nil, fmt.Errorf("getSavedRaidFiles(): %w", err)
	}

	// Add files to the file list
	for _, file := range savedRaidsFileInfo {
		if strings.Contains(file.Name(), "RaidAttend") {
			fileName := file.Name()
			//fmt.Printf("Found Raid Dump file: %s...\n", fileName)
			savedRaidsFileList = append(savedRaidsFileList, fileName)
		}
	}

	//fmt.Printf("Returning a fileList with %d files from inside (%s)...\n", len(raidDumpFileList), basePath)
	return savedRaidsFileList, nil
}

func getSavedRaidModDate(raidFilePath string) (string, error) {
	// Get the directory of the current executable
	EQpath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getSavedRaidModDate(): os.getwd: %w", err)
	}
	// Get the directory of the SavedRaids folder
	raidLogsFolder := EQpath + "\\SavedRaids"

	fileStat, err := os.Stat(raidLogsFolder + "\\" + raidFilePath)
	if err != nil {
		return "", fmt.Errorf("getSavedRaidModDate(%s): os.Stat: %w", raidFilePath, err)
	}
	fileModified := fileStat.ModTime().String()
	return fileModified, nil
}

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

func playerIsInCache(name string) bool {
	for _, player := range core.GetActivePlayers() {
		if player.Name == name {
			return true
		}
	}
	return false
}

func AddPlayersToRaid() {
	// Ensure all players in core.Players are also in ActiveRaid.Players
	for _, player := range core.Players {
		if !PlayerIsInRaid(player.Name) {
			fmt.Printf("Adding %s to raid...\n", player.Name)
			ActiveRaid.Players = append(ActiveRaid.Players, player)
		}
	}
}

func PlayerIsInDisplayList(characterName string) bool {
	// Check if the provided characterName is in the DisplayList map
	_, ok := DisplayList[characterName]
	return ok
}

// Updates the DisplayList map with cached player information, enforcing alias
func UpdateDisplayList() {
	// Clear the DisplayList map
	DisplayList = make(map[string]int)
	// Iterate through the checkins map
	for character := range ActiveRaid.Checkins {
		handle := alias.TryToGetHandle(character)
		// If the handle of the checkin character is not in the DisplayList map
		if !PlayerIsInDisplayList(handle) {
			// Add the handle of the checkin character to the DisplayList map
			DisplayList[handle] = GetHighestCheckin(handle)
			fmt.Printf("Adding %s to DisplayList...\n", handle)
		}
	}
}

// Return the highest checkin associated with the handle specified
func GetHighestCheckin(handle string) int {
	highestCheckin := 0
	for character, checkins := range ActiveRaid.Checkins {
		// Check if the character is associated with the provided handler
		if alias.TryToGetHandle(character) == handle {
			if checkins > highestCheckin {
				highestCheckin = checkins
			}
		}
	}
	return highestCheckin
}

// Detirmine if the provided character name is present in the active raid
func PlayerIsInRaid(characterName string) bool {
	for _, player := range ActiveRaid.Players {
		if player.Name == characterName {
			return true
		}
	}
	return false
}

// Saves the provided raid to a json file
func SaveRaid(raid Raid) error {
	fmt.Printf("Saving (%s) to raid file...\n", raid.Name)

	EQpath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("organizeRaidDumps(): os.getwd: %w", err)
	}

	// Get SavedRaids folder
	savedRaidsFolder := EQpath + "\\SavedRaids"

	file, err := json.MarshalIndent(raid, "", " ")
	if err != nil {
		return fmt.Errorf("SaveRaid(): failed to marshal raid: %w", err)
	}

	//fmt.Println(string(file))

	err = ioutil.WriteFile(savedRaidsFolder+"\\"+raid.FileName, file, 0644)

	if err != nil {
		return fmt.Errorf("SaveRaid(): failed to write to raid file: %w", err)
	}

	fmt.Println("Raid save successful!")
	return nil
}

// Load the provided raid from a json file
func LoadRaid(fileName string) (Raid, error) {
	fmt.Printf("Loading (%s) from raid file...\n", fileName)

	if fileName == "" {
		return Raid{}, fmt.Errorf("LoadRaid(): no file name provided")
	}

	EQpath, err := os.Getwd()
	if err != nil {
		return Raid{}, fmt.Errorf("LoadRaid(): os.getwd: %w", err)
	}

	// Get SavedRaids folder
	savedRaidsFolder := EQpath + "\\SavedRaids"

	file, err := ioutil.ReadFile(savedRaidsFolder + "\\" + fileName)
	if err != nil {
		return Raid{}, fmt.Errorf("LoadRaid(): failed to read raid file: %w", err)
	}

	var raid Raid
	err = json.Unmarshal(file, &raid)
	if err != nil {
		return Raid{}, fmt.Errorf("LoadRaid(): failed to unmarshal raid: %w", err)
	}

	fmt.Println("Raid load successful!")
	return raid, nil
}
