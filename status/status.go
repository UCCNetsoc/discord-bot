package status

import (
	"fmt"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/commands"
	"github.com/bwmarrin/discordgo"
)

// Status ...
func Status(s *discordgo.Session) {
	for {
		minecraftPlayerCount(s)
		wait := time.After(time.Minute)
		<-wait
	}
}

func minecraftPlayerCount(s *discordgo.Session) {
	resp, err := commands.Query()
	if err != nil {
		log.Error("Failed to query MC Server status: " + err.Error())
		time.Sleep(time.Hour)
	} else {
		plural := "players"
		if resp.Players.Online == 1 {
			plural = "player"
		}
		s.UpdateStatus(0, fmt.Sprintf("Minecraft %d %s online minecraft.netsoc.co", resp.Players.Online, plural))
	}
}
