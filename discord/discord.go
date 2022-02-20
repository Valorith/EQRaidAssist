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

func SendEmbedMessage(title, description string, messageType int) error {
	var err error
	var color uint32 = 3583291
	name := "EQRaidAssist"
	url := "https://www.clumsysworld.com/"
	icon_url := "https://styles.redditmedia.com/t5_2rosz/styles/communityIcon_hvzzme5v9kz41.jpg"
	if messageType == 1 { // Loot Channel
		discordwh.WebhookURL, err = config.GetLootWebHookUrl()
		if err != nil {
			return fmt.Errorf("discord: failed to get loot webhook url: %s", err)
		}
	} else if messageType == 2 { // Attendance Channel
		discordwh.WebhookURL, err = config.GetAtendWebHookUrl()
		if err != nil {
			return fmt.Errorf("discord: failed to get attendance webhook url: %s", err)
		}
	}
	author := discordwh.Author{Name: name, URL: url, IconURL: icon_url}
	embed := discordwh.Embed{Author: &author, Title: title, URL: url, Description: description, Color: color, Fields: nil, Thumbnail: nil, Image: nil, Footer: nil}
	PO := discordwh.PostOptions{Username: name, AvatarURL: icon_url, Embeds: []discordwh.Embed{embed}}
	err = discordwh.Post(PO)
	if err != nil {
		return fmt.Errorf("discord: failed to post embed: %v", err)
	}
	return nil
}
