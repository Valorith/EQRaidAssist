package core

import (
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
