package commands

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func upcomingEvent(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	calendarURL := viper.GetString("google.calendar.public.ics")
	if i.GuildID == viper.GetString("discord.committee.server") {
		calendarURL = viper.GetString("google.calendar." + i.ApplicationCommandData().Options[0].StringValue() + ".ics")
	}
	eventEmbeds, err := upcomingEventEmbeds(ctx, s, 2, calendarURL)
	if err != nil {
		log.WithContext(ctx).WithError(err)
		InteractionResponseError(s, i, err.Error(), true)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: eventEmbeds,
		},
	})
	if err != nil {
		log.WithContext(ctx).WithError(err)
	}
}

func upcomingEventEmbeds(ctx context.Context, s *discordgo.Session, limit int, url string) (eventEmbeds []*discordgo.MessageEmbed, err error) {
	upcomingEvents, err := api.QueryCalendarEvents(url)
	if err != nil {
		return nil, err
	}
	if len(upcomingEvents) < 1 {
		return nil, errors.New("There are currently no events scheduled, Stay tuned!")
	}
	for i, event := range upcomingEvents {
		if i == limit {
			break
		}
		emb := embed.NewEmbed()
		emb.SetTitle(event.Summary)

		if len(event.Description) > 0 {
			emb.SetDescription(strings.ReplaceAll(event.Description, `\n`, "\n"))
		}
		if len(event.Location) > 0 {
			emb.AddField("Where?", event.Location)
		}
		// Parse fileID from calendar event url, then use it as a param on google drive url to get raw image
		re := regexp.MustCompile(`\/file\/d\/([^\/]+)`)
		for _, attachment := range event.Attachments {
			if attachment.Mime[:5] == "image" {
				emb.SetImage(fmt.Sprintf(" https://drive.google.com/uc?id=%s", re.FindString(attachment.Value)[8:]))
			}
		}
		if emb.Image == nil && viper.GetString("google.calendar.image.default") != "" {
			emb.SetThumbnail(viper.GetString("google.calendar.image.default"))
		}

		emb.AddField("When?", fmt.Sprintf("<t:%v:F>", event.Start.Unix()))

		emb.SetAuthor("Netsoc Event", s.State.User.AvatarURL("2048"), "https://netsoc.co/go/calendar")

		eventEmbeds = append(eventEmbeds, emb.MessageEmbed)
	}
	return eventEmbeds, nil
}
