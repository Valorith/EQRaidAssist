package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

var (
	// Public variables
	Token       string
	BotPrefix   string
	LootChannel string
	WebHookUrl  string
	// Private variables
	config *configStruct
	mu     sync.RWMutex
)

type configStruct struct {
	Token       string `json:"Token"`
	BotPrefix   string `json:"BotPrefix"`
	LootChannel string `json:"LootChannel"`
	WebHookUrl  string `json:"WebHookUrl"`
}

func GetBotToken() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	if Token == "" {
		return "", fmt.Errorf("bot token not set")
	}
	return Token, nil
}

func SetBotToken(token string) error {
	mu.RLock()
	defer mu.RUnlock()
	if token == "" {
		return fmt.Errorf("SetWebHookUrl(): provided token is invalid")
	}
	config.Token = token
	Token = token
	err := SaveConfig()
	if err != nil {
		return fmt.Errorf("SetBotToken(): %w", err)
	}
	return nil
}

func GetBotPrefix() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	if BotPrefix == "" {
		return "", fmt.Errorf("bot prefix not set")
	}
	return BotPrefix, nil
}

func SetBotPrefix(prefix string) error {
	mu.RLock()
	defer mu.RUnlock()
	if prefix == "" {
		return fmt.Errorf("SetWebHookUrl(): provided prefix is invalid")
	}
	config.BotPrefix = prefix
	BotPrefix = prefix
	err := SaveConfig()
	if err != nil {
		return fmt.Errorf("SetBotToken(): %w", err)
	}
	return nil
}

func GetLootChannel() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	if LootChannel == "" {
		return "", fmt.Errorf("loot channel not set")
	}
	return LootChannel, nil
}

func SetLootChannel(channelID string) error {
	mu.RLock()
	defer mu.RUnlock()
	if channelID == "" {
		return fmt.Errorf("SetWebHookUrl(): provided channel id is invalid")
	}
	config.LootChannel = channelID
	LootChannel = channelID
	err := SaveConfig()
	if err != nil {
		return fmt.Errorf("SetBotToken(): %w", err)
	}
	return nil
}

func GetWebHookUrl() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	if WebHookUrl == "" {
		return "", fmt.Errorf("web hook url not set")
	}
	return WebHookUrl, nil
}

func SetWebHookUrl(url string) error {
	mu.RLock()
	defer mu.RUnlock()
	if url == "" {
		return fmt.Errorf("SetWebHookUrl(): provided url is invalid")
	}
	config.WebHookUrl = url
	WebHookUrl = url
	err := SaveConfig()
	if err != nil {
		return fmt.Errorf("SetBotToken(): %w", err)
	}
	return nil
}

func ReadConfig() error {
	fmt.Println("Reading config file...")

	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	//fmt.Println(string(file))

	err = json.Unmarshal(file, &config)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix
	LootChannel = config.LootChannel
	WebHookUrl = config.WebHookUrl

	if err == nil {
		fmt.Println("Config load successful!")
	}

	return nil

}

func SaveConfig() error {
	fmt.Println("Saving to config file...")

	file, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return fmt.Errorf("SaveConfig(): failed to marshal config: %w", err)
	}

	err = ioutil.WriteFile("./config.json", file, 0644)
	if err != nil {
		return fmt.Errorf("SaveConfig(): failed to write to config: %w", err)
	}

	fmt.Println("Config save successful!")

	return nil

}
