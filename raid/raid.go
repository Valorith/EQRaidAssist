package raid

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Valorith/EQRaidAssist/core"
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
	ActiveRaid = Raid{
		Name:        "RaidAttend_" + strconv.Itoa(currentYear) + "_" + strconv.Itoa(int(currentMonth)) + "_" + strconv.Itoa(currenteDay) + "_" + strconv.Itoa(currentHour) + "_" + strconv.Itoa(currentMinute) + "_" + strconv.Itoa(currentSecond),
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
	return nil
}

func (raid Raid) CheckIn() {

	//Increment checkinCounts
	for playerName, checkIns := range raid.Checkins {
		//Ensure player is on the current player list
		if playerStillActive(playerName) {
			raid.Checkins[playerName] = checkIns + 1
		}
	}
}

func (raid Raid) PrintParticipation() error {
	if !raid.Active {
		return fmt.Errorf("Raid is not active")
	}
	if len(raid.Players) == 0 {
		return fmt.Errorf("there are no players in the raid")
	}
	fmt.Println("Print Raid: " + raid.Name)
	//Increment checkinCounts
	for playerName, checkIns := range raid.Checkins {
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
	for _, player := range ActiveRaid.Players {
		if player.Name == name {
			return true
		}
	}
	return false
}
