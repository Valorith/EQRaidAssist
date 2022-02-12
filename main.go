package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	loadFile "github.com/Valorith/EQRaidAssist/loadFIle"
)

func main() {
	players := loadPlayers() // loads players from the newest Raid Dump file in the EQ directory
	playerCount := len(players)
	if playerCount > 0 {
		fmt.Printf("%d Players Succesfully loaded.\n", playerCount)
	} else {
		fmt.Println("Player Loading Failed!")
	}

	for _, player := range players {
		player.PrintPlayer()
	}
}

type Player struct {
	Name  string
	Level int
	Class string
	Group int
	Loot  []string
}

func loadPlayers() []Player {
	EQpath, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	//logsFolder := EQpath + "\\Logs"
	fmt.Println("Loading Players from: ", EQpath)
	raidDumpFileList := []string{}
	fileListIndex := 0
	filePathError := filepath.Walk(EQpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("Scanning File: %s....\n", path)
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
	}

	oldestIndex := 0
	//newestIndex := len(raidDumpFileList) - 1
	loadFileDirectory := raidDumpFileList[oldestIndex]
	fmt.Println("Newest Raid Dump File Detected: ", loadFileDirectory)
	var dumpLines []string = loadFile.Load(loadFileDirectory)
	var players []Player

	for _, line := range dumpLines {
		formattedLine := strings.Replace(line, "\t", ",", -1)
		groupNumber, _ := strconv.Atoi(formattedLine[0:strings.Index(formattedLine, ",")])
		formattedLine = formattedLine[strings.Index(formattedLine, ",")+1:]
		charName := formattedLine[0:strings.Index(formattedLine, ",")]
		formattedLine = formattedLine[strings.Index(formattedLine, ",")+1:]
		charLevel, _ := strconv.Atoi(formattedLine[0:strings.Index(formattedLine, ",")])
		formattedLine = formattedLine[strings.Index(formattedLine, ",")+1:]
		charClass := formattedLine[0:strings.Index(formattedLine, ",")]
		fmt.Println("Player Detected: ", charName)
		players = append(players, Player{charName, charLevel, charClass, groupNumber, []string{}})
	}

	return players
}

func (player *Player) PrintPlayer() {

	fmt.Println("Char Name: ", player.Name)
	fmt.Println("Char Level: ", player.Level)
	fmt.Println("Char Class: ", player.Class)
	fmt.Println("Group Number: ", player.Group)
	fmt.Println("Loot:")
	for _, lootItem := range player.Loot {
		fmt.Println("\t", lootItem)
	}
	fmt.Println("-----------------")

}

func (player *Player) AddLoot(lootItem string) {
	if lootItem != "" {
		player.Loot = append(player.Loot, lootItem)
	} else {
		fmt.Println("Error adding loot item: ", lootItem)
	}
}
