package commands

import (
	"context"
	"fmt"

	"github.com/UCCNetsoc/discord-bot/embed"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
)

const layoutIE = "02/01/06"

func ping(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "pong")
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to send pong message")
		return
	}
}

func help(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	emb := embed.NewEmbed()
	emb.SetTitle("Netsoc Bot Commands")
	description := ""
	for k, v := range helpStrings {
		description += fmt.Sprintf("**`!%s`**: %s\n", k, v)
	}
	if isCommittee(s, m) {
		description += "\n**Committee commands**:\n\n"
		for k, v := range committeeHelpStrings {
			description += fmt.Sprintf("**`!%s`**: %s\n", k, v)
		}
	}
	emb.SetDescription(description)
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, emb.MessageEmbed)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to send help message")
		return
	}
}
