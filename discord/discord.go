package discord

import (
	"fmt"

	"github.com/Valorith/EQRaidAssist/config"
	"github.com/bwmarrin/discordgo"
)

var BotID string
var lootChannelID string
var goBot *discordgo.Session
var err error

func Start() {
	token := config.Token // Discord bot token
	if token == "" {
		fmt.Println("No bot token found in config.json")
		return
	}
	goBot, err = discordgo.New("Bot " + token)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")

	if err != nil {
		fmt.Println(err.Error())
	}

	BotID = u.ID

	goBot.AddHandler(messageHandler)

	err = goBot.Open()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is running!")
	<-make(chan struct{})
}

func Stop() error {
	err := goBot.Close()
	if err != nil {
		return fmt.Errorf("Stop(): failed to close discord session: %w", err)
	}
	fmt.Println("Bot has shut down...")
	return nil
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == BotID {
		return
	}

	if m.Content == "ping" {
		_, err = s.ChannelMessageSend(m.ChannelID, "pong")
	}
	if err != nil {
		fmt.Println("messageHandler:", err.Error())
	}
}

func SendLootMessage(m string) {

	_, err = goBot.ChannelMessageSend(config.LootChannel, m)
	if err != nil {
		fmt.Println("messageHandler:", err.Error())
	}
}
