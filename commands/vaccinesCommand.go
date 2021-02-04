package commands

import (
	"context"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/corona"
	"github.com/bwmarrin/discordgo"
)

func vaccines(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	vaccines, err := corona.GetVaccines()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Error querying vaccines from arcgis")
		return
	}
	s.ChannelMessageSendEmbed(m.ChannelID, vaccines.Embed(nil))
}
