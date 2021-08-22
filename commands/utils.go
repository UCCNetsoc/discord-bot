package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/embed"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
)

func errorEmbed(message string) *discordgo.MessageEmbed {
	return embed.NewEmbed().SetTitle("❗ ERROR ❗").SetDescription(message).MessageEmbed
}

func RefeshSchedule(schedule *gocron.Scheduler, s *discordgo.Session) {
	schedule.Clear()
	upcomingEvents, err := api.QueryFacebookEvents()
	if err != nil {
		log.WithError(err).Error("Error occured refrshing cron scheduler event")
	}
	if len(upcomingEvents) > 0 {
		schedule.Every(1).Hour().StartAt((time.Unix((upcomingEvents[0].Date - 600), 0))).Do(UpcomingEventAnnounce, context.TODO(), s)
	}
}

func InteractionResponseError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Encountered error: %v", err),
			Flags:   1 << 6,
		},
	})
}
