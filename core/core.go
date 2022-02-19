package core

import (
	"fmt"

	"github.com/Valorith/EQRaidAssist/player"
)

var (
	Players []*player.Player // Players detected within the raid dump file
)

func GetActivePlayers() []*player.Player {
	return Players
}

func AddPlayer(p *player.Player) {
	Players = append(Players, p)
}

func ClearPlayers() {
	Players = nil
	fmt.Println("Cached players cleared...")
}
