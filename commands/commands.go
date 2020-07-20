package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/embed"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

const layoutIE = "02/01/06"

// ping command
func ping(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "pong")
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to send pong message")
		return
	}
}

// help command
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

// register command
func serverRegister(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	if _, ok := registering[m.Author.ID]; ok {
		return
	}

	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("failed to create DM channel")
		s.ChannelMessageSendEmbed(m.ChannelID, errorEmbed("Failed to create a private message channel."))
		return
	}

	registering[m.Author.ID] = initiatedRegistration

	emb := embed.NewEmbed().
		SetTitle("Netsoc Server Registration").
		SetDescription("Send me your UCC email address so we can verify you're a UCC student.").
		SetFooter("Message your email in the form <Student ID>@umail.ucc.ie. A code will be sent to your email that you will then send here.")
	s.ChannelMessageSendEmbed(channel.ID, emb.MessageEmbed)
}

func serverJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	ctx := context.WithValue(context.Background(), log.Key, log.Fields{
		"user_id":  m.User.ID,
		"guild_id": m.GuildID,
	})

	servers := viper.Get("discord.servers").(*config.Servers)
	publicServer, err := s.Guild(servers.PublicServer)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to get Public Server guild")
		return
	}
	if m.GuildID != publicServer.ID {
		return
	}
	// Handle join messages
	messages := *viper.Get("discord.welcome_messages").(*[]string)
	if len(messages) > 0 {
		i := rand.Intn(len(messages))
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Couldnt find guild for welcome")
			return
		}
		welcomeID := guild.SystemChannelID
		if len(welcomeID) > 0 {
			// Send welcome message
			emb := embed.NewEmbed().SetTitle("Welcome!").SetDescription(fmt.Sprintf(messages[i], m.Member.Mention()))
			// s.ChannelMessageSend(welcomeID, fmt.Sprintf(messages[i], m.Member.Mention()))
			if viper.GetBool("discord.autoregister") {
				emb.SetFooter("We've sent you a DM so you can register for full access to the server!\nIf you're a student in another college simply let us know here and we will be able to assign you a role manually!")
			} else {
				emb.SetFooter("Please type `!register` to start the verification process to make sure you're a UCC student.\nIf you're a student in another college simply let us know here and we will be able to assign you a role manually!")
			}

			s.ChannelMessageSendEmbed(welcomeID, emb.MessageEmbed)
		}

	}

	if viper.GetBool("discord.autoregister") {
		// Handle users joining by auto registering them
		serverRegister(ctx, s, &discordgo.MessageCreate{Message: &discordgo.Message{Author: m.User}})
		return
	}
}

func addEvent(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	event(ctx, s, m, "@everyone")
}

func addEventSilent(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	event(ctx, s, m, "everyone")
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
		event.ImgData = b
		if len(event.Description) < viper.GetInt("discord.charlimit") {
			s.MessageReactionAdd(m.ChannelID, m.ID, string(twitter))
			reactionMap[m.ID] = event
		}

	} else {
		s.ChannelMessageSend(m.ChannelID, "This command is unavailable")
	}
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
						return
					}
				}
			}
		}
	}
}
