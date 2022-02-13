package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	loadFile "github.com/Valorith/EQRaidAssist/loadFIle"
)

//Global Variables
var lastLoadedRaidFile string = ""

func main() {
	LoadedRaidFile, err := getNewestRaidFile()
	lastLoadedRaidFile = LoadedRaidFile
	if err != nil {
		fmt.Println("lastLoadedRaidFile:", err)
	}
	fScanner := FileScanner{LoadedRaidFile, false}
	//Load Players in
	players, err := loadPlayers(lastLoadedRaidFile) // loads players from the newest Raid Dump file in the EQ directory
	if err != nil {
		fmt.Println("Player load failed: ", err)
	}
	playerCount := len(players)
	if playerCount > 0 {
		fmt.Printf("%d Players Succesfully loaded.\n", playerCount)
	} else {
		fmt.Println("Player Loading Failed!")
	}

	for _, player := range players {
		player.PrintPlayer()
	}

	//Get User user input
	getUserInput(&fScanner)
}

type Player struct {
	Name  string
	Level int
	Class string
	Group int
	Loot  []string
}

type FileScanner struct {
	fileLocation string
	enabled      bool
}

func getUserInput(scanner *FileScanner) {
	for {
		fmt.Printf("Commands:\nStart scanning raid file: 'start' or 'run'\nStop scanning raid file: 'stop'\nExit application: 'exit' or 'quit'\n")
		fmt.Println("-----------------")
		fmt.Println("Enter a command:")
		var userInput string
		fmt.Scanln(&userInput)
		switch userInput {
		case "start":
			err := scanner.Start()
			if err != nil {
				fmt.Println("Error starting scanner: ", err)
			}
		case "run":
			if !scanner.enabled {
				if scanner.fileLocation != "" {
					go scanner.scan()
				} else {
					fmt.Println("Error: No raid file location detected!")
				}
			}
		case "stop":
			if scanner.enabled {
				fmt.Println("[Status] Stopping Scanner...")
				scanner.enabled = false
			} else {
				fmt.Println("The file scanner is not running.")
			}
		case "exit":
			fmt.Println("[Status] Exiting...")
			os.Exit(0)
		case "quit":
			fmt.Println("[Status] Exiting...")
			os.Exit(0)
		default:
			fmt.Println("[Status] Invalid command.")
		}
	}
}

func getNewestRaidFile() (string, error) {
	EQpath, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("getwd: %w", err)
	}
	//logsFolder := EQpath + "\\Logs"
	//fmt.Println("Loading Players from: ", EQpath)
	raidDumpFileList := []string{}
	fileListIndex := 0
	filePathError := filepath.Walk(EQpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("filepath.Walk: %w", err)
		}
		//fmt.Printf("Scanning File: %s....\n", path)
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
		return "", fmt.Errorf("filepath.Walk: %w", filePathError)
	}

	newestIndex := len(raidDumpFileList) - 1
	return raidDumpFileList[newestIndex], nil
}

func loadPlayers(fileLocation string) ([]Player, error) {
	fmt.Println("Newest Raid Dump File Detected: ", fileLocation)
	var dumpLines []string = loadFile.Load(fileLocation)
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

	return players, nil
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

func (scanner *FileScanner) scan() error {
	// Ensure Scanner fileLocation is set
	if scanner.fileLocation == "" {
		return nil
	}
	if !scanner.IsRunning() {
		scanner.enabled = true // Ensure scanner is enabled
	}
	for {
		if scanner.IsRunning() {
			fmt.Println("Scanning File: ", scanner.fileLocation)
			loadedRaidFile, err := getNewestRaidFile()
			if err != nil {
				return err
			}

			if loadedRaidFile != lastLoadedRaidFile {
				//Load Players in
				players, err := loadPlayers(loadedRaidFile) // loads players from the newest Raid Dump file in the EQ directory
				if err != nil {
					fmt.Println("Player load failed: ", err)
				}
				playerCount := len(players)
				if playerCount > 0 {
					fmt.Printf("%d Players Succesfully loaded.\n", playerCount)
				} else {
					fmt.Println("Player Loading Failed!")
				}

				for _, player := range players {
					player.PrintPlayer()
				}
				lastLoadedRaidFile = loadedRaidFile
			} else {
				fmt.Println("No new raid file detected.")
			}
			time.Sleep(time.Second * 15)
		} else {
			return nil
		}
	}
}

func (scanner *FileScanner) IsRunning() bool {
	return scanner.enabled
}

func (scanner *FileScanner) Start() error {
	if !scanner.IsRunning() {
		if scanner.fileLocation != "" {
			go scanner.scan()
		} else {
			return fmt.Errorf("no raid file location detected")
		}
	} else {
		return fmt.Errorf("scanner is already running")
	}
	return nil
}
