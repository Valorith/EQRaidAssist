package main

import (
	"fmt"
	"os"

	"github.com/Valorith/EQRaidAssist/scanner"
)

func main() {
	var err error

	for {
		var userInput string

		if !scanner.IsServerNameSet() {
			fmt.Printf("Enter the server short name: ")
			fmt.Scanln(&userInput)
			err = scanner.SetServerName(userInput)
			if err != nil {
				fmt.Println("invalid server name:", err)
				continue
			}
		}

		if !scanner.IsCharacterNameSet() {
			fmt.Printf("Enter your characters first name: ")
			fmt.Scanln(&userInput)
			err = scanner.SetCharacterName(userInput)
			if err != nil {
				fmt.Println("invalid character name:", err)
				continue
			}
		}

		fmt.Printf("Commands:\nStart scanning raid file: 'start' or 'run'\nStop scanning raid file: 'stop'\nExit application: 'exit' or 'quit'\n")
		fmt.Println("-----------------")
		fmt.Println("Enter a command:")

		fmt.Scanln(&userInput)
		err = getUserInput(userInput)
		if err != nil {
			fmt.Printf("failed %s: %s\n", userInput, err)
		}
	}

}

func getUserInput(input string) error {
	var err error

	switch input {
	case "start":
		err = scanner.Start()
	case "run":
		err = scanner.Start()
	case "stop":
		scanner.Stop()
	case "timer":
		scanner.SetRaidFrequency(16)
	case "exit":
		fmt.Println("[Status] Exiting...")
		os.Exit(0)
	case "quit":
		fmt.Println("[Status] Exiting...")
		os.Exit(0)
	default:
		return fmt.Errorf("invalid command")
	}
	return err
}
