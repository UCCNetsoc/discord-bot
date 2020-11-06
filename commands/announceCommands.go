package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/UCCNetsoc/discord-bot/prometheus"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func addEvent(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	event(ctx, s, m, "@everyone")
}

func addEventSilent(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	event(ctx, s, m, "everyone")
}

func addEventWebsite(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	channels := viper.Get("discord.channels").(*config.Channels)
	if isCommittee(s, m) && m.ChannelID == channels.PrivateEvents {
		_, err := api.ParseEvent(m, committeeHelpStrings["event"])
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("failed to parse event")
			s.ChannelMessageSend(m.ChannelID, "Failed to parse event: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event successfully posted to website! (Depending on cache may take a few minutes)")
	} else {
		s.ChannelMessageSend(m.ChannelID, "This command is unavailable")
	}
}

func event(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, mention string) {
	channels := viper.Get("discord.channels").(*config.Channels)
	if isCommittee(s, m) && m.ChannelID == channels.PrivateEvents {
		event, err := api.ParseEvent(m, committeeHelpStrings["event"])
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("failed to parse event")
			s.ChannelMessageSend(m.ChannelID, "Failed to parse event: "+err.Error())
			return
		}
		b := bytes.NewBuffer([]byte{})
		s.ChannelFileSendWithMessage(
			channels.PublicAnnouncements,
			fmt.Sprintf(
				"Hey %s, we have another upcoming event on *%s*:\n**%s**\n%s",
				mention,
				event.Date.Format(layoutIE),
				event.Title,
				event.Description,
			),
			"poster.jpg",
			io.TeeReader(event.ImgData, b),
		)
		prometheus.EventCreate()
		event.ImgData = b
		if len(event.Description) < viper.GetInt("discord.charlimit") {
			s.MessageReactionAdd(m.ChannelID, m.ID, string(twitter))
			reactionMap[m.ID] = event
		}

	} else {
		s.ChannelMessageSend(m.ChannelID, "This command is unavailable")
	}
}

func upcomingEvent(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, mention string) {
	upcomingEvents, err := api.QueryFacebookEvents()
	nearest := upcomingEvents[0]
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error occured parsing upcoming event")
		log.WithError(err).WithContext(ctx).Error("Error occured parsing upcoming event")
	}
	title := "Netsoc Upcoming Event"
	p := message.NewPrinter(language.English)
	body := p.Sprintf("**%s**\n", nearest.Title)
	body += p.Sprintf("%s\n", nearest.Description)

	body += p.Sprintf("**When?**\n%s\n", time.Unix(nearest.Date, 0).Format("Jan 2 at 3:04 PM"))

	emb := embed.NewEmbed()
	emb.SetColor(0xc20002)
	emb.SetTitle(title)
	emb.SetDescription(body)
	emb.SetImage(nearest.ImageURL)
	s.ChannelMessageSendEmbed(m.ChannelID, emb.MessageEmbed)
}

func addAnnouncement(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	announcement(ctx, s, m, "@everyone\n")
}

func addAnnouncementSilent(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	announcement(ctx, s, m, "")
}

func announcement(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, mention string) {
	channels := viper.Get("discord.channels").(*config.Channels)
	if isCommittee(s, m) && m.ChannelID == channels.PrivateEvents {
		announcement, err := api.ParseAnnouncement(m, committeeHelpStrings["announce"])
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("error sending announcement")
			s.ChannelMessageSend(m.ChannelID, "Error sending announcement: "+err.Error())
			return
		}
		if announcement.ImgData != nil {
			b := bytes.NewBuffer([]byte{})
			s.ChannelFileSendWithMessage(
				channels.PublicAnnouncements,
				fmt.Sprintf("%s%s", mention, announcement.Content),
				"poster.jpg",
				io.TeeReader(announcement.ImgData, b),
			)
			announcement.ImgData = b
		} else {
			s.ChannelMessageSend(channels.PublicAnnouncements, fmt.Sprintf("%s%s", mention, announcement.Content))
		}
		if len(announcement.Content) < viper.GetInt("discord.charlimit") {
			s.MessageReactionAdd(m.ChannelID, m.ID, string(twitter))
			reactionMap[m.ID] = announcement
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "This command is unavailable")
	}
}

// recall events and announcements
func recall(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	channels := viper.Get("discord.channels").(*config.Channels)
	if isCommittee(s, m) && m.ChannelID == channels.PrivateEvents {
		public, err := s.ChannelMessages(channels.PublicAnnouncements, 100, "", "", "")
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Error getting channel public")
		}
		private, err := s.ChannelMessages(channels.PrivateEvents, 100, "", "", "")
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Error getting channel private")
		}
		for _, message := range private {
			if strings.HasPrefix(message.Content, viper.GetString("bot.prefix")+"announce"+" ") {
				content := strings.TrimPrefix(message.Content, viper.GetString("bot.prefix")+"announce"+" ")
				s.ChannelMessageDelete(channels.PrivateEvents, message.ID)
				for _, publicMessage := range public {
					publicContent := strings.Trim(strings.Join(strings.Split(publicMessage.Content, "\n")[1:], "\n"), " ")
					if publicContent == content {
						s.ChannelMessageDelete(channels.PublicAnnouncements, publicMessage.ID)
						s.ChannelMessageSend(m.ChannelID, "Successfully recalled announcement\n*"+publicContent+"*")
						return
					}
				}

			} else if strings.HasPrefix(message.Content, viper.GetString("bot.prefix")+"sannounce"+" ") {
				content := strings.TrimPrefix(message.Content, viper.GetString("bot.prefix")+"sannounce"+" ")
				s.ChannelMessageDelete(channels.PrivateEvents, message.ID)
				for _, publicMessage := range public {
					publicContent := strings.Trim(publicMessage.Content, " ")
					if publicContent == content {
						s.ChannelMessageDelete(channels.PublicAnnouncements, publicMessage.ID)
						s.ChannelMessageSend(m.ChannelID, "Successfully recalled announcement\n*"+publicContent+"*")
						return
					}
				}
			} else if strings.HasPrefix(message.Content, viper.GetString("bot.prefix")+"event"+" ") {
				create := &discordgo.MessageCreate{Message: message}
				event, err := api.ParseEvent(create, committeeHelpStrings["event"])
				if err != nil {
					log.WithContext(ctx).WithError(err).Error("failed to parse event")
					continue
				}
				// Found event
				s.ChannelMessageDelete(channels.PrivateEvents, message.ID)
				content := fmt.Sprintf(
					"Hey @everyone, we have a new upcoming event on *%s*:\n**%s**\n%s",
					event.Date.Format(layoutIE),
					event.Title,
					event.Description,
				)
				for _, publicMessage := range public {
					if content == publicMessage.Content {
						s.ChannelMessageDelete(channels.PublicAnnouncements, publicMessage.ID)
						s.ChannelMessageSend(m.ChannelID, "Successfully recalled event\n"+fmt.Sprintf("**%s**\n%s", event.Title, event.Description))
						prometheus.EventRevoke()
						return
					}
				}
			}
		}
	}
}
