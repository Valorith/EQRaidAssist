package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Valorith/EQRaidAssist/alias"
	"github.com/Valorith/EQRaidAssist/config"
	"github.com/Valorith/EQRaidAssist/raid"
	"github.com/Valorith/EQRaidAssist/scanner"
)

var (
	commandsDisplayed bool = false
)

func main() {
	defer close()
	var err error
	var count int //scanline arg return count
	var userInput string

	err = config.ReadConfig()
	if err != nil {
		fmt.Printf("main: failed to read config: %s\n", err)
	}
	err = alias.ReadAliases()
	if err != nil {
		fmt.Printf("main: failed to read aliases: %s\n", err)
	}
	for {
		// If the character name is not set, request it
		if !scanner.IsCharacterNameSet() {
			fmt.Printf("Enter the character name that you want to monitor (first name only): ")
			count, err = fmt.Scanln(&userInput)
			if err != nil {
				fmt.Println("invalid name input:", err)
				continue
			}
			if count != 1 {
				fmt.Println("incorrect number of arguments for character name")
				continue
			}
			err = scanner.SetCharacterName(strings.Title(userInput))
			if err != nil {
				fmt.Println("invalid character name:", err)
				continue
			}
		}

		// If the server name is not set, attempt to infer it
		if !scanner.IsServerNameSet() {
			inferServerName()
		}

		// If the server name is not set (infer failed), request it
		if !scanner.IsServerNameSet() {
			fmt.Printf("Enter your server short name: ")
			count, err = fmt.Scanln(&userInput)
			if err != nil {
				fmt.Println("invalid server input:", err)
				continue
			}
			if count != 1 {
				fmt.Println("incorrect number of arguments for server name")
				continue
			}
			err = scanner.SetServerName(userInput)
			if err != nil {
				fmt.Println("invalid server name:", err)
				continue
			}
		}

		if !commandsDisplayed {
			// Print the available commands to the user
			printCommands()
		}

		var subCommand, value string
		_, err = fmt.Scanln(&userInput, &subCommand, &value)
		if err != nil {
			if !(err.Error() == "unexpected newline") { // Filter out this specific known error condition that is not of concern
				fmt.Printf("command error: %s %s %s: %s\n", userInput, subCommand, value, err)
			}
		}
		// Retrieve commands from user
		go getUserInput(userInput, subCommand, value)
	}
}

// Print the available commands to the user
func printCommands() {
	commandsDisplayed = true
	fmt.Printf("Commands:\nStart scanning raid file: 'start'\nStop scanning raid file: 'stop'\nExit application: 'exit' or 'quit'\n")
	fmt.Printf("Load the most recent saved raid: 'get lastraid'\n")
	fmt.Printf("Show current raid participants: 'get raid'\n")
	fmt.Printf("Reset all session data: 'set reset'\n")
	fmt.Printf("Get app variables: 'get <identifier>'\nSet app variables: 'set <identifier>'\n")
	fmt.Println("-----------------")
	fmt.Println("Enter a command:")
}

// Attempt to infer what the server name is based on the character name and client files
func inferServerName() {
	// Attempt to infer the server name based upon the provided char name
	fmt.Println("Attempting to infer server name...")
	possibleServerNames, err := config.GetPossibleServerNames(scanner.GetCharacterName())
	if err != nil {
		fmt.Println("main: failed to load possible server names:", err)
	}
	if len(possibleServerNames) > 1 {
		var err error
		var count int //scanline arg return count
		var userInput string
		index := 1
		for _, serverName := range possibleServerNames {
			fmt.Printf("%d) %s\n", index, serverName)
			index++
		}
		// Set the server to the selected server
		fmt.Printf("Enter your server selection (1 - %d):\n", len(possibleServerNames))
		count, err = fmt.Scanln(&userInput)
		if err != nil {
			fmt.Println("invalid server input:", err)
			return
		}
		selection, err := strconv.Atoi(userInput)
		if err != nil {
			fmt.Println("strconv.Atoi:", err)
			return
		}
		if count != 1 {
			fmt.Println("incorrect number of arguments for server name")
			return
		}
		selectedServer := possibleServerNames[selection-1]
		fmt.Println("Setting server to:", selectedServer)
		err = scanner.SetServerName(selectedServer)
		if err != nil {
			fmt.Println("invalid server name:", err)
			return
		}
	} else if len(possibleServerNames) == 1 {
		err = scanner.SetServerName(possibleServerNames[0])
		if err != nil {
			fmt.Println("main: failed to set server name:", err)
		}
		fmt.Printf("Setting server to: %s (infered)\n", possibleServerNames[0])
	} else {
		fmt.Println("main: failed to infer server name")
	}
}

func getUserInput(input, subcommand, value string) {
	var err error
	fmt.Println("-----------------")
	// Primary Command Handler
	switch input {
	case "start":
		if scanner.IsRunning() {
			fmt.Printf("getUserInput: scanner is already running")
			return
		}
		scanner.Start()
	case "stop":
		scanner.Stop()
	case "exit":
		fmt.Println("[Status] Exiting...")
		config.SaveConfig()
		os.Exit(0)
	case "set":
		switch subcommand {
		case "server":
			fmt.Println("Setting server name to:", value)
			err = scanner.SetServerName(value)
			if err != nil {
				fmt.Printf("getUserInput: invalid server name: %s", err)
			}
		case "character":
			fmt.Println("Setting character name to:", value)
			err = scanner.SetCharacterName(value)
			if err != nil {
				fmt.Printf("getUserInput: invalid character name: %s->%s", value, err)
			}

			if scanner.IsRunning() {
				scanner.Reboot()
			}
		case "token":
			fmt.Println("Setting bot token to:", value)
			err = config.SetBotToken(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s\n", value, err)
			}
		case "prefix":
			fmt.Println("Setting bot prefix to:", value)
			err = config.SetBotPrefix(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s\n", value, err)
			}
		case "channel":
			fmt.Println("Setting loot channel to:", value)
			err = config.SetLootChannel(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s\n", value, err)
			}
		case "lootwebhook":
			fmt.Println("Setting loot webhook url to:", value)
			err = config.SetLootWebHookUrl(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s\n", value, err)
			}
		case "attendwebhook":
			fmt.Println("Setting attendance webhook url to:", value)
			err = config.SetAttendWebHookUrl(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s\n", value, err)
			}
		case "guild":
			fmt.Println("Setting guild name to:", value)
			err = config.SetGuildName(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s\n", value, err)
			}
		case "guildalias":
			fmt.Println("Importing the guild list from file and generating the alias list...")
			err := alias.GenerateAliasListFromGuildList()
			if err != nil {
				fmt.Printf("GenerageAliasListFromGuildList(): %s\n", err)
			}
		case "timer":
			fmt.Println("Setting timer to:", value)
			// convert string to int
			intValue, err := strconv.Atoi(value)
			if err != nil {
				fmt.Printf("getUserInput: invalid timer value: %s\n", err)
			}
			scanner.SetRaidFrequency(intValue)
		case "raid":
			if value == "checkin" {
				err := raid.ActiveRaid.CheckIn()
				if err != nil {
					fmt.Printf("ActiveRaid.CheckIn(): %s\n", err)
				}
			} else {
				fmt.Printf("ActiveRaid.CheckIn(): %s\n", "invalid subcommand")
			}
		case "reboot":
			if scanner.IsRunning() {
				scanner.Reboot()
			} else {
				fmt.Println("The scanner is not running...")
			}
		case "reset":
			/*
				scanner.Stop()
				core.ClearPlayers()
				scanner.ResetData()
				discordwh.ResetData()
				raid.ResetData()
				config.ResetData()
			*/
			fmt.Println("Reset Not yet implemented")
		default:
			fmt.Printf("invalid command(%s)\n", subcommand)
		}
	case "get":
		switch subcommand {
		case "server":
			serverName := scanner.GetServerName()
			fmt.Println("Server Name:", serverName)
		case "character":
			characterName := scanner.GetCharacterName()
			fmt.Println("Character Name:", characterName)
		case "token":
			token, err := config.GetBotToken()
			if err != nil {
				fmt.Printf("GetBotToken(): %s\n", err)
			}
			fmt.Println("Bot Token:", token)
		case "prefix":
			prefix, err := config.GetBotPrefix()
			if err != nil {
				fmt.Printf("GetBotPrefix(): %s\n", err)
			}
			fmt.Println("Bot Prefix:", prefix)
		case "channel":
			channel, err := config.GetLootChannel()
			if err != nil {
				fmt.Printf("GetLootChannel(): %s\n", err)
			}
			fmt.Println("Bot Loot Channel:", channel)
		case "lootwebhook":
			webHookUrl, err := config.GetLootWebHookUrl()
			if err != nil {
				fmt.Printf("GetLootWebHookUrl(): %s\n", err)
			}
			fmt.Println("Loot Web Hook Url:", webHookUrl)
		case "attendwebhook":
			webHookUrl, err := config.GetAtendWebHookUrl()
			if err != nil {
				fmt.Printf("GetAtendWebHookUrl(): %s\n", err)
			}
			fmt.Println("Attendance Web Hook Url:", webHookUrl)
		case "guild":
			guildName, err := config.GetGuildName()
			if err != nil {
				fmt.Printf("GetGuildName(): %s\n", err)
			}
			fmt.Println("Guild Name:", guildName)
		case "guildalias":
			err := alias.ReadGuildMembers()
			if err != nil {
				fmt.Printf("ReadGuildMembers(): %s\n", err)
			}
			fmt.Println("Guild aliases loaded...")
		case "timer":
			timer := scanner.GetRaidTimer()
			fmt.Println("Raid file scan timer:", timer)
		case "raid":
			switch value {
			case "":
				err := raid.ActiveRaid.PrintParticipation()
				if err != nil {
					fmt.Printf("ActiveRaid.PrintParticipation(): %s\n", err)
				}
			default:
				fmt.Printf("raid: invalid subcommand --> %s\n", subcommand)
			}
		case "alias":
			switch value {
			case "":
				err := alias.ActiveAliases.PrintAliases()
				if err != nil {
					fmt.Printf("ActiveAliases.PrintAliases(): %s\n", err)
				}
			default:
				fmt.Printf("alias: invalid subcommand --> %s\n", subcommand)
			}
		case "displaylist":
			switch value {
			case "":
				err := raid.PrintDisplayList()
				if err != nil {
					fmt.Printf("raid.PrintDisplayList(): %s\n", err)
				}
			default:
				fmt.Printf("displaylist: invalid subcommand --> %s\n", subcommand)
			}
		case "lastraid":
			err := raid.ActiveRaid.Load(value)
			if err != nil {
				fmt.Printf("ActiveRaid.Load(): %s\n", err)
			}
		case "commands":
			printCommands()
		case "quit":
			fmt.Println("[Status] Exiting...")
			config.SaveConfig()
			os.Exit(0)
		default:
			fmt.Printf("invalid command(%s)\n", subcommand)
		}
	case "alias":
		characterName := subcommand
		handle := value
		if characterName == "" || handle == "" {
			fmt.Println("invalid command: Expected: alias <characterName> <handle>")
			return
		}
		err := alias.AddAlias(characterName, handle)
		if err != nil {
			fmt.Printf("AddAlias(): %s\n", err)
		}
	case "ping":
		fmt.Println("Pong")
	default:
		fmt.Printf("invalid command(%s)\n", input)
	}
	if err != nil {
		fmt.Printf("failed command: %s %s %s: %s\n", input, subcommand, value, err)
	}
}

func close() {
	fmt.Println("Cleaning up before exit...")
	config.SaveConfig()
}
