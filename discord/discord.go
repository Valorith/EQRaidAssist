package discord

import (
	"github.com/Valorith/EQRaidAssist/discordwh"
)

func SendLootMessage(m string) {

	discordwh.WebhookURL = "https://discord.com/api/webhooks/944311340909084693/hG2bzbr7BkczxPaDqRBqQY-o19z8WLcRcXkola07rzqUflvex4v-7CTkrtjXs4ufUTB-"
	discordwh.Say(m)

}
