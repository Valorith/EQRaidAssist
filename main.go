package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Valorith/EQRaidAssist/config"
	"github.com/Valorith/EQRaidAssist/scanner"
)

func main() {
	var err error
	var count int //scanline arg return count
	var userInput string

	err = config.ReadConfig()
	if err != nil {
		fmt.Printf("main: failed to read config: %s", err)
	}

	for {
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

		if !scanner.IsServerNameSet() {
			inferServerName()
		}

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

		printCommands()
		var subCommand, value string
		_, err = fmt.Scanln(&userInput, &subCommand, &value)
		if err != nil {
			if !(err.Error() == "unexpected newline") { // Filter out this specific known error condition that is not of concern
				fmt.Printf("command error: %s %s %s: %s\n", userInput, subCommand, value, err)
			}
		}
		//write a for loop to handle multiple commands
		go getUserInput(userInput, subCommand, value)
	}
}

func printCommands() {
	fmt.Printf("Commands:\nStart scanning raid file: 'start'\nStop scanning raid file: 'stop'\nExit application: 'exit' or 'quit'\n")
	fmt.Printf("Get app variables: 'get <identifier>'\nSet app variables: 'set <identifier>'\n")
	fmt.Println("-----------------")
	fmt.Println("Enter a command:")
}

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
		case "webhook":
			fmt.Println("Setting webhook url to:", value)
			err = config.SetWebHookUrl(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s\n", value, err)
			}
		case "timer":
			fmt.Println("Setting timer to:", value)
			// convert string to int
			intValue, err := strconv.Atoi(value)
			if err != nil {
				fmt.Printf("getUserInput: invalid timer value: %s\n", err)
			}
			scanner.SetRaidFrequency(intValue)
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
		case "webhook":
			webHookUrl, err := config.GetWebHookUrl()
			if err != nil {
				fmt.Printf("GetWebHookUrl(): %s\n", err)
			}
			fmt.Println("Web Hook Url:", webHookUrl)
		case "timer":
			timer := scanner.GetRaidTimer()
			fmt.Println("Raid file scan timer:", timer)
		case "ping":
			fmt.Println("Pong")
		case "quit":
			fmt.Println("[Status] Exiting...")
			os.Exit(0)
		default:
			fmt.Println("invalid command")
		}
	}
	if err != nil {
		fmt.Printf("failed command: %s %s %s: %s\n", input, subcommand, value, err)
	}
}
