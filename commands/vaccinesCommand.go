package commands

import (
	"context"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/corona"
	"github.com/bwmarrin/discordgo"
)

func vaccines(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	vaccines, err := corona.GetVaccines()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Error querying vaccines from arcgis")
		return
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{vaccines.Embed(nil)},
		},
	})
	if err != nil {
		log.WithContext(ctx).WithError(err)
	}
}
