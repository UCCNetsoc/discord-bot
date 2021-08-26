package commands

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func upcomingEvent(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	eventEmbeds, err := upcomingEventEmbeds(ctx, s, -1)
	if err != nil {
		log.WithContext(ctx).WithError(err)
		InteractionResponseError(s, i, err.Error(), true)
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

func upcomingEventEmbeds(ctx context.Context, s *discordgo.Session, limit int) (eventEmbeds []*discordgo.MessageEmbed, err error) {
	upcomingEvents := api.QueryCalendarEvents()
	if len(upcomingEvents) < 1 {
		return nil, errors.New("There are currently no events scheduled, Stay tuned!")
	}
	p := message.NewPrinter(language.English)
	for i, event := range upcomingEvents {
		if i == limit {
			break
		}
		emb := embed.NewEmbed()
		emb.SetTitle(event.Summary)

		if len(event.Description) > 0 {
			emb.SetDescription(p.Sprintf("%s\n", event.Description))
		}
		if len(event.Location) > 0 {
			emb.AddField("Where?", event.Location)
		}
		// Parse fileID from calendar event url, then use it as a param on google drive url to get raw image
		re := regexp.MustCompile(`\/file\/d\/([^\/]+)`)
		for _, attachment := range event.Attachments {
			if attachment.Mime[:5] == "image" {
				// TODO: Needs testing to see if this always works, else get raw image url from drive url redirect header
				emb.SetImage(fmt.Sprintf(" https://drive.google.com/uc?id=%s", re.FindString(attachment.Value)[8:]))
			}
		}
		if emb.Image == nil {
			emb.SetImage("https://avatars.githubusercontent.com/u/6690158?s=400&u=fb42911a2e865d716c137619afdf1f7a266989cf&v=4")
		}

		emb.AddField("When?", fmt.Sprintf("<t:%v:F>", event.Start.Unix()))

		emb.SetAuthor("Netsoc Event", s.State.User.AvatarURL("2048"), "https://netsoc.co/go/calendar")

		eventEmbeds = append(eventEmbeds, emb.MessageEmbed)
	}
	fmt.Printf("\n%+v\n\n", eventEmbeds)
	return eventEmbeds, nil
}
