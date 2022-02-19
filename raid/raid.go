package raid

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Valorith/EQRaidAssist/core"
	"github.com/Valorith/EQRaidAssist/discord"
	"github.com/Valorith/EQRaidAssist/player"
)

var (
	Active     bool
	ActiveRaid Raid
)

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
	Active      bool             `json:"active"`      // Indicates whether the raid is active or not
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
		Name:        "RaidAttend_" + strconv.Itoa(currentYear) + strconv.Itoa(int(currentMonth)) + strconv.Itoa(currenteDay) + "-" + strconv.Itoa(currentHour) + "-" + strconv.Itoa(currentMinute),
		StartYear:   currentYear,
		StartMonth:  int(currentMonth),
		StartDay:    currenteDay,
		StartHour:   currentHour,
		StartMinute: currentMinute,
		StartSecond: currentSecond,
		Description: "",
		Checkins:    make(map[string]int),
		Players:     activePlayers,
		Active:      true}
	//-----------------------
	ActiveRaid.initializeCheckins()
	return nil
}

func (raid Raid) initializeCheckins() {
	for _, player := range raid.Players {
		ActiveRaid.Checkins[player.Name] = 1
	}
	//fmt.Printf("%d checkins initialized at a count of 1...\n", len(ActiveRaid.Checkins))
}

func (raid Raid) CheckIn() error {
	discord.SendMessage("[Raid Atendance] Raid Checkin Initiated!", 2)
	//Increment checkinCounts
	for playerName, checkIns := range raid.Checkins {
		//Ensure player is on the current player list
		if playerStillActive(playerName) {
			raid.Checkins[playerName] = checkIns + 1
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
	for playerName, checkIns := range ActiveRaid.Checkins {
		//fmt.Printf("%s has checked in %d times\n", playerName, checkIns)
		//Ensure player is on the current player list
		if playerStillActive(playerName) {
			fmt.Printf("%s: %d [Active]\n", playerName, checkIns)
		} else {
			fmt.Printf("%s: %d [Inactive]\n", playerName, checkIns)
		}
	}
	return nil
}

func playerStillActive(name string) bool {
	for _, player := range core.GetActivePlayers() {
		if player.Name == name {
			return true
		}
	}
	return false
}

func PlayerIsInRaid(player player.Player) bool {
	for _, p := range ActiveRaid.Players {
		if p.Name == player.Name {
			//fmt.Printf("%s is already in the raid...\n", player.Name)
			return true
		}
	}
	return false
}

func AddPlayersToRaid() {
	for _, player := range core.GetActivePlayers() {
		if !PlayerIsInRaid(*player) {
			fmt.Printf("Adding %s to raid...\n", player.Name)
			discord.SendMessage("[Raid Attendance] "+player.Name+" has joined the raid!", 2)
			ActiveRaid.Players = append(ActiveRaid.Players, player)
		}
	}
}
