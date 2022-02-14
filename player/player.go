package player

import (
	"fmt"
	"strconv"
	"strings"
)

type Player struct {
	Name  string
	Level int
	Class string
	Group int
	Loot  []string
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
		out = fmt.Sprintf("%s\t %s", out, lootItem)
	}
	out = fmt.Sprintf("%s\n------------------\n", out)
	return out
}

func (p *Player) AddLoot(lootItem string) {
	if lootItem != "" {
		p.Loot = append(p.Loot, lootItem)
	} else {
		fmt.Println("Error adding loot item: ", lootItem)
	}
}
