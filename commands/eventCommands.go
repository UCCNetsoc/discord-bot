package commands

import (
	"context"
	"errors"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func upcomingEvent(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	eventEmbed, err := upcomingEventEmbed(ctx, s)
	if err != nil {
		InteractionResponseError(s, i, err)
		return
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{eventEmbed},
		},
	})
	if err != nil {
		log.WithContext(ctx).WithError(err)
	}
}

func UpcomingEventAnnounce(ctx context.Context, s *discordgo.Session) {
	channels := viper.Get("discord.channels").(*config.Channels)
	eventEmbed, _ := upcomingEventEmbed(ctx, s)
	s.ChannelMessageSend(channels.PublicAnnouncements, "@here\nEvent starting in 10 minutes!\n\n")
	s.ChannelMessageSendEmbed(channels.PublicAnnouncements, eventEmbed)
}

func upcomingEventEmbed(ctx context.Context, s *discordgo.Session) (eventEmbed *discordgo.MessageEmbed, err error) {
	upcomingEvents, err := api.QueryFacebookEvents()
	if len(upcomingEvents) < 1 {
		return nil, errors.New("There are currently no events scheduled, Stay tuned!")
	}
	title := "Netsoc Upcoming Event"
	emb := embed.NewEmbed()
	emb.SetColor(0xc20002)
	emb.SetTitle(title)
	p := message.NewPrinter(language.English)
	body := ""
	nearest := upcomingEvents[0]
	if err != nil {
		log.WithError(err).WithContext(ctx).Error("Error occured parsing upcoming event")
		return nil, errors.New("error occured parsing upcoming event")
	}
	body += p.Sprintf("**%s**\n", nearest.Title)
	body += p.Sprintf("%s\n", nearest.Description)
	body += p.Sprintf("**When?**\n%s\n", time.Unix(nearest.Date, 0).Format("Jan 2 at 3:04 PM"))
	emb.SetImage(nearest.ImageURL)
	return emb.MessageEmbed, nil
}