package commands

import (
	"context"

	"github.com/spf13/viper"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
)

func ping(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "pong",
		}})
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to send pong message")
		return
	}
}

func version(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: viper.GetString("bot.version"),
		}})
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to send version message")
		return
	}
}
