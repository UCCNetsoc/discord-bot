package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func upcomingEvent(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	eventEmbed, err := upcomingEventEmbed(ctx, s)
	if err != nil {
		log.WithContext(ctx).WithError(err)
		InteractionResponseError(s, i, err.Error(), true)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  1 << 6,
			Embeds: []*discordgo.MessageEmbed{eventEmbed},
		},
	})
	if err != nil {
		log.WithContext(ctx).WithError(err)
	}
}

// TODO: Look into the possibility of parsing images from the ics data, else consider a default banner to better fill out the embed
func upcomingEventEmbed(ctx context.Context, s *discordgo.Session) (eventEmbed *discordgo.MessageEmbed, err error) {
	upcomingEvents := api.QueryCalendarEvents()
	if len(upcomingEvents) < 1 {
		return nil, errors.New("There are currently no events scheduled, Stay tuned!")
	}
	emb := embed.NewEmbed()
	emb.SetTitle(upcomingEvents[0].Summary)

	p := message.NewPrinter(language.English)
	if len(upcomingEvents[0].Description) > 0 {
		emb.SetDescription(p.Sprintf("%s\n", upcomingEvents[0].Description))

	}
	emb.AddField("When?", fmt.Sprintf("<t:%v:F>", upcomingEvents[0].Start.Unix()))
	return emb.MessageEmbed, nil
}
