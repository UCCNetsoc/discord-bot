package commands

import (
	"fmt"

	"github.com/UCCNetsoc/discord-bot/embed"

	"github.com/bwmarrin/discordgo"
)

func errorEmbed(message string) *discordgo.MessageEmbed {
	return embed.NewEmbed().SetTitle("❗ ERROR ❗").SetDescription(message).MessageEmbed
}

// TODO: Reconsider if cron scheduler is still required, bot event announcements were never really used

func InteractionResponseError(s *discordgo.Session, i *discordgo.InteractionCreate, errorMessage string, tagError bool) {
	if tagError {
		errorMessage = fmt.Sprintf("Encountered error: %v", errorMessage)
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: errorMessage,
			Flags:   1 << 6,
		},
	})
}
