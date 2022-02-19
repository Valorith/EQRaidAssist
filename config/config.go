package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	// Public variables
	Token            string
	BotPrefix        string
	LootChannel      string
	LootWebHookUrl   string
	AttendWebHookUrl string
	// Private variables
	config *configStruct
	mu     sync.RWMutex
)

type configStruct struct {
	Token            string `json:"Token"`
	BotPrefix        string `json:"BotPrefix"`
	LootChannel      string `json:"LootChannel"`
	LootWebHookUrl   string `json:"LootWebHookUrl"`
	AttendWebHookUrl string `json:"AttendWebHookUrl"`
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
		return fmt.Errorf("SetBotToken(): provided token is invalid")
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
		return fmt.Errorf("SetBotPrefix(): provided prefix is invalid")
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
		return fmt.Errorf("SetLootChannel(): provided channel id is invalid")
	}
	config.LootChannel = channelID
	LootChannel = channelID
	err := SaveConfig()
	if err != nil {
		return fmt.Errorf("SetBotToken(): %w", err)
	}
	return nil
}

func GetLootWebHookUrl() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	if LootWebHookUrl == "" {
		return "", fmt.Errorf("loot web hook url not set")
	}
	return LootWebHookUrl, nil
}

func GetAtendWebHookUrl() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	if AttendWebHookUrl == "" {
		return "", fmt.Errorf("attendance web hook url not set")
	}
	return AttendWebHookUrl, nil
}

func SetLootWebHookUrl(url string) error {
	mu.RLock()
	defer mu.RUnlock()
	if url == "" {
		return fmt.Errorf("SetLootWebHookUrl(): provided url is invalid")
	}
	config.LootWebHookUrl = url
	LootWebHookUrl = url
	err := SaveConfig()
	if err != nil {
		return fmt.Errorf("SetLootWebHookUrl(): %w", err)
	}
	return nil
}

func SetAttendWebHookUrl(url string) error {
	mu.RLock()
	defer mu.RUnlock()
	if url == "" {
		return fmt.Errorf("SetAttendWebHookUrl(): provided url is invalid")
	}
	config.AttendWebHookUrl = url
	AttendWebHookUrl = url
	err := SaveConfig()
	if err != nil {
		return fmt.Errorf("SetAttendWebHookUrl(): %w", err)
	}
	return nil
}

func GetPossibleServerNames(charName string) ([]string, error) {
	EQpath, err := os.Getwd()
	if err != nil {
		return []string{}, fmt.Errorf("os.Getwd(): %w", err)
	}
	//logsFolder := EQpath + "\\Logs"
	//fmt.Println("Loading Players from: ", EQpath)
	charFileList := []string{}
	filePathError := filepath.Walk(EQpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("filepath.Walk: %w", err)
		}
		//fmt.Printf("Scanning File: %s....\n", path)
		isDir := info.IsDir()
		fileName := info.Name()
		comparisonString := charName + "_"
		if !isDir {
			//fmt.Printf("Scanning File: %s; comparing to: %s\n", fileName, comparisonString)
			if strings.Contains(fileName, comparisonString) {
				fileServerName := fileName[strings.LastIndex(fileName, "_")+1 : strings.LastIndex(fileName, ".")]
				if !sliceContains(charFileList, fileServerName) {
					charFileList = append(charFileList, fileServerName)
				}
			}
		}
		return nil
	})

	if filePathError != nil {
		fmt.Println(filePathError)
		return []string{}, fmt.Errorf("filepath.Walk: %w", filePathError)
	}

	return charFileList, nil
}

func sliceContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
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
	LootWebHookUrl = config.LootWebHookUrl
	AttendWebHookUrl = config.AttendWebHookUrl

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
