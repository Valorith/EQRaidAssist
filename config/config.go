package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	// Public variables
	Token       string
	BotPrefix   string
	LootChannel string
	// Private variables
	config *configStruct
)

type configStruct struct {
	Token       string `json:"Token"`
	BotPrefix   string `json:"BotPrefix"`
	LootChannel string `json:"LootChannel"`
}

func ReadConfig() error {
	fmt.Println("Reading config file...")

	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println(string(file))

	err = json.Unmarshal(file, &config)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix
	LootChannel = config.LootChannel

	return nil
}
