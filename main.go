package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/Valorith/EQRaidAssist/scanner"
)

var (
	mu sync.RWMutex
)

func main() {
	var err error
	var count int //scanline arg return count
	for {
		var userInput string

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

		fmt.Printf("Commands:\nStart scanning raid file: 'start' or 'run'\nStop scanning raid file: 'stop'\nExit application: 'exit' or 'quit'\n")
		fmt.Println("-----------------")
		fmt.Println("Enter a command:")
		var subCommand, value string
		_, err = fmt.Scanln(&userInput, &subCommand, &value)
		if err != nil {
			if !(err.Error() == "unexpected newline") { // Filter out this specific known error condition that is not of concern
				fmt.Printf("command error: %s %s %s: %s\n", userInput, subCommand, value, err)
				continue
			}
		}
		//write a for loop to handle multiple commands
		err = getUserInput(userInput, subCommand, value)
		if err != nil {
			fmt.Printf("failed command: %s %s %s: %s\n", userInput, subCommand, value, err)
			continue
		}
	}
}

func getUserInput(input, subcommand, value string) error {
	mu.Lock()
	defer mu.Unlock()
	var err error

	// Primary Command Handler
	switch input {
	case "start":
		err = scanner.Start()
	case "run":
		err = scanner.Start()
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
				return fmt.Errorf("getUserInput: invalid server name: %s", err)
			}
		case "character":
			fmt.Println("Setting character name to:", value)
			err = scanner.SetCharacterName(value)
			if err != nil {
				return fmt.Errorf("getUserInput: invalid character name: %s->%s", value, err)
			}
		case "timer":
			fmt.Println("Setting timer to:", value)
			// convert string to int
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("getUserInput: invalid timer value: %s", err)
			}
			scanner.SetRaidFrequency(intValue)
		}
	case "get":
		switch subcommand {
		case "server":
			serverName := scanner.GetServerName()
			fmt.Println("Server Name:", serverName)
		}
	case "quit":
		fmt.Println("[Status] Exiting...")
		os.Exit(0)
	default:
		return fmt.Errorf("invalid command")
	}
	return err
}
