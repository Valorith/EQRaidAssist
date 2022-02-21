package core

import (
	"fmt"

	"github.com/Valorith/EQRaidAssist/alias"
	"github.com/Valorith/EQRaidAssist/player"
)

var (
	Players []*player.Player // Players detected within the raid dump file
)

func GetActivePlayers() []*player.Player {
	return Players
}

func AddPlayer(p *player.Player) error {
	handle := alias.TryToGetHandle(p.Name)
	if IsCachedPlayer(handle) {
		return fmt.Errorf("player %s is already cached", handle)
	}
	Players = append(Players, p)
	return nil
}

func ClearPlayers() {
	Players = nil
	fmt.Println("Cached players cleared...")
}

// Check if the provided characterName is in the list of players
func IsCachedPlayer(characterName string) bool {
	for _, p := range Players {
		if p.Name == characterName {
			return true
		}
	}
	return false
}
