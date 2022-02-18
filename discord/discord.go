package discord

import (
	"fmt"

	"github.com/Valorith/EQRaidAssist/config"
	"github.com/Valorith/EQRaidAssist/discordwh"
)

func SendLootMessage(m string) {
	var err error
	discordwh.WebhookURL, err = config.GetWebHookUrl()
	if err != nil {
		fmt.Println("discord: failed to get webhook url:", err)
		return
	}
	discordwh.Say(m)

}
