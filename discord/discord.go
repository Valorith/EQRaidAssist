package discord

import (
	"fmt"

	"github.com/Valorith/EQRaidAssist/config"
	"github.com/Valorith/EQRaidAssist/discordwh"
)

func SendMessage(m string, messageType int) {
	var err error
	if messageType == 1 { // Loot Channel
		discordwh.WebhookURL, err = config.GetLootWebHookUrl()
	} else if messageType == 2 { // Attendance Channel
		discordwh.WebhookURL, err = config.GetAtendWebHookUrl()
	}

	if err != nil {
		fmt.Println("discord: failed to get webhook url:", err)
		return
	}
	discordwh.Say(m)

}
