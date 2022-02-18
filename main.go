package main

import (
	"fmt"
	"os"
	"strconv"

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

		if !scanner.IsCharacterNameSet() {
			fmt.Printf("Enter your characters first name: ")
			count, err = fmt.Scanln(&userInput)
			if err != nil {
				fmt.Println("invalid name input:", err)
				continue
			}
			if count != 1 {
				fmt.Println("incorrect number of arguments for character name")
				continue
			}
			err = scanner.SetCharacterName(userInput)
			if err != nil {
				fmt.Println("invalid character name:", err)
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
				fmt.Printf("getUserInput: %s->%s", value, err)
			}
		case "prefix":
			fmt.Println("Setting bot prefix to:", value)
			err = config.SetBotPrefix(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s", value, err)
			}
		case "channel":
			fmt.Println("Setting loot channel to:", value)
			err = config.SetLootChannel(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s", value, err)
			}
		case "webhook":
			fmt.Println("Setting webhook url to:", value)
			err = config.SetWebHookUrl(value)
			if err != nil {
				fmt.Printf("getUserInput: %s->%s", value, err)
			}
		case "timer":
			fmt.Println("Setting timer to:", value)
			// convert string to int
			intValue, err := strconv.Atoi(value)
			if err != nil {
				fmt.Printf("getUserInput: invalid timer value: %s", err)
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
				fmt.Printf("GetBotToken(): %s", err)
			}
			fmt.Println("Bot Token:", token)
		case "prefix":
			prefix, err := config.GetBotPrefix()
			if err != nil {
				fmt.Printf("GetBotPrefix(): %s", err)
			}
			fmt.Println("Bot Prefix:", prefix)
		case "channel":
			channel, err := config.GetLootChannel()
			if err != nil {
				fmt.Printf("GetLootChannel(): %s", err)
			}
			fmt.Println("Bot Loot Channel:", channel)
		case "webhook":
			webHookUrl, err := config.GetWebHookUrl()
			if err != nil {
				fmt.Printf("GetWebHookUrl(): %s", err)
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
