package player

import (
	"fmt"
	"strconv"
	"strings"
)

// Represents an EverQuest player
type Player struct {
	Name  string     `json:"name"`     // Name of the player
	Level int        `json:"level"`    // Level of the player
	Class string     `json:"class"`    // Class of the player
	Group int        `json:"group"`    // Raid Group number
	Loot  []LootItem `json:"lootitem"` // Loot attributed to the player
}

type LootItem struct {
	Name        string `json:"name"`
	Count       int    `json:"count"`
	Description string `json:"description"`
}

// NewFromLine takes a line argument and creates a new player
func NewFromLine(line string) (*Player, error) {
	var err error
	p := &Player{}

	formattedLine := strings.Replace(line, "\t", ",", -1)
	in := formattedLine[0:strings.Index(formattedLine, ",")]
	p.Group, err = strconv.Atoi(in)
	if err != nil {
		return nil, fmt.Errorf("atoi groupNumber %s: %w", in, err)
	}
	formattedLine = formattedLine[strings.Index(formattedLine, ",")+1:]
	p.Name = formattedLine[0:strings.Index(formattedLine, ",")]
	formattedLine = formattedLine[strings.Index(formattedLine, ",")+1:]
	in = formattedLine[0:strings.Index(formattedLine, ",")]
	p.Level, err = strconv.Atoi(in)
	if err != nil {
		return nil, fmt.Errorf("atoi charLevel %s: %w", in, err)
	}
	formattedLine = formattedLine[strings.Index(formattedLine, ",")+1:]
	p.Class = formattedLine[0:strings.Index(formattedLine, ",")]
	return p, nil
}

func (p *Player) String() string {

	out := fmt.Sprintf("Char Name: %s", p.Name)
	out = fmt.Sprintf("%s\nChar Level: %d", out, p.Level)
	out = fmt.Sprintf("%s\nChar Class: %s", out, p.Class)
	out = fmt.Sprintf("%s\nGroup Number: %d", out, p.Group)
	out = fmt.Sprintf("%s\nLoot: ", out)
	for _, lootItem := range p.Loot {
		out = fmt.Sprintf("%s\t %s\n", out, lootItem.Name)
	}
	out = fmt.Sprintf("%s\n------------------\n", out)
	return out
}

func (p *Player) AddLoot(item LootItem) error {
	if item.Name != "" {
		p.Loot = append(p.Loot, item)
	} else {
		fmt.Println("Error adding loot item: ", item)
		return fmt.Errorf("player: AddLoot: Error adding loot item: %s to player (%s)", item.Name, p.Name)
	}
	return nil
}
