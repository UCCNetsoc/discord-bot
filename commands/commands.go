package commands

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/config"
	petname "github.com/dustinkirkland/golang-petname"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var registering = make([]string, 0)
var verifyCodes = make(map[string]string)

const layoutIE = "02/01/06"

// ping command
func ping(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "pong")
	if err != nil {
		log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Failed to send pong message")
		return
	}
}

// help command
func help(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	out := "```"
	for k, v := range helpStrings {
		out += k + ": " + v + "\n"
	}
	if isCommittee(m) {
		for k, v := range committeeHelpStrings {
			out += k + ": " + v + "\n"
		}
	}
	_, err := s.ChannelMessageSend(m.ChannelID, out+"```")
	if err != nil {
		log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Failed to send help message")
		return
	}
}

// register command
func serverRegister(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, a := range registering {
		if a == m.Author.ID {
			return
		}
	}
	registering = append(registering, m.Author.ID)

	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Failed to create DM channel")
		return
	}

	s.ChannelMessageSend(channel.ID, "Please message me your UCC email address so I can verify you as a member of UCC")
}

func serverJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	ctx := context.WithValue(context.Background(), logKey, log.Fields{
		"user_id":  m.User.ID,
		"guild_id": m.GuildID,
	})

	servers := viper.Get("discord.servers").(*config.Servers)
	publicServer, err := s.Guild(servers.PublicServer)
	if err != nil {
		log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Failed to get Public Server guild")
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
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Couldnt find guild for welcome")
			return
		}
		welcomeID := guild.SystemChannelID
		if len(welcomeID) > 0 {
			// Send welcome message
			s.ChannelMessageSend(welcomeID, fmt.Sprintf(messages[i], m.Member.Mention()))
			if viper.GetBool("discord.autoregister") {
				s.ChannelMessageSend(welcomeID, "We've sent you a DM so you can register for full access to the server!\nIf you're a student in another college simply let us know here and we will be able to assign you a role manually!")
			} else {
				s.ChannelMessageSend(welcomeID, "Please type `!register` to start the verification process to make sure you're a UCC student.\nIf you're a student in another college simply let us know here and we will be able to assign you a role manually!")
			}
		}

	}

	if viper.GetBool("discord.autoregister") {
		// Handle users joining by auto registering them
		serverRegister(ctx, s, &discordgo.MessageCreate{Message: &discordgo.Message{Author: m.User}})
		return
	}
}

func addEvent(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	channels := viper.Get("discord.channels").(*config.Channels)
	if isCommittee(m) && m.ChannelID == channels.PrivateEvents {
		event, err := api.ParseEvent(m, committeeHelpStrings["event"])
		if err != nil {
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("failed to parse event")
			s.ChannelMessageSend(m.ChannelID, "Failed to parse event: "+err.Error())
			return
		}
		s.ChannelFileSendWithMessage(
			channels.PublicAnnouncements,
			fmt.Sprintf(
				"Hey @everyone, we have a new upcoming event on *%s*:\n**%s**\n%s",
				event.Date.Format(layoutIE),
				event.Title,
				event.Description,
			),
			"poster.jpg",
			event.Image.Body,
		)

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
	if isCommittee(m) && m.ChannelID == channels.PrivateEvents {
		announcement, err := api.ParseAnnouncement(m, committeeHelpStrings["announce"])
		if err != nil {
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("error sending announcement")
			s.ChannelMessageSend(m.ChannelID, "Error sending announcement: "+err.Error())
			return
		}

		if announcement.Image != nil {
			s.ChannelFileSendWithMessage(
				channels.PublicAnnouncements,
				fmt.Sprintf("%s%s", mention, announcement.Content),
				"poster.jpg",
				announcement.Image.Body,
			)
		} else {
			s.ChannelMessageSend(channels.PublicAnnouncements, fmt.Sprintf("%s%s", mention, announcement.Content))
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "This command is unavailable")
	}
}

// recall events and announcements
func recall(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	channels := viper.Get("discord.channels").(*config.Channels)
	if isCommittee(m) && m.ChannelID == channels.PrivateEvents {
		public, err := s.ChannelMessages(channels.PublicAnnouncements, 100, "", "", "")
		if err != nil {
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Error getting channel public")
		}
		private, err := s.ChannelMessages(channels.PrivateEvents, 100, "", "", "")
		if err != nil {
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Error getting channel private")
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
					log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("failed to parse event")
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

func quote(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	var mention *discordgo.User
	if len(m.Mentions) > 0 {
		mention = m.Mentions[0]
	}
	servers := viper.Get("discord.servers").(*config.Servers)

	allChannels, err := s.GuildChannels(servers.PublicServer)
	if err != nil {
		log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Couldn't find public guild")
		return
	}
	blacklist := *viper.Get("discord.quote_blacklist").(*[]string)
	attempts := 0
	if len(allChannels) == 0 {
		log.WithFields(ctx.Value(logKey).(log.Fields)).Error(fmt.Sprintf("Got no channels for user %v ", *mention))
		s.ChannelMessageSend(m.ChannelID, "Couldn't find any messages by that user")
		return
	}
	max := len(allChannels) / 2
	for len(allChannels) > 0 || attempts > max {
		channels := []*discordgo.Channel{}
		// Get all text channels
		for _, channel := range allChannels {
			if channel != nil {
				block := false
				for _, blocked := range blacklist {
					if channel.ID == blocked {
						block = true
					}
				}
				if !block {
					perms, err := s.UserChannelPermissions(s.State.User.ID, channel.ID)
					if err != nil {
						log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Error getting channel perms")
						return
					}
					if channel.Type == discordgo.ChannelTypeGuildText &&
						perms&discordgo.PermissionReadMessages > 0 {
						channels = append(channels, channel)
					}
				}
			}
		}
		if len(channels) == 0 {
			log.WithFields(ctx.Value(logKey).(log.Fields)).Error("Error getting messages for: " + m.Author.Username)
			s.ChannelMessageSend(m.ChannelID, "Couldn't find any messages by that user")
			return
		}
		// Check if channel mentioned
		for _, channel := range channels {
			if strings.Contains(m.Content, channel.Mention()) {
				channels = []*discordgo.Channel{channel}
				break
			}
		}
		channelIndex := rand.Intn(len(channels))
		channel := channels[channelIndex]

		discMessages := &Ring{}
		// Get cached messages
		if cachedMessages != nil {
			if more, exists := cachedMessages.Get(channel.ID); exists {
				discMessages = more.(*Ring)
				if discMessages.Len() == 0 {
					continue
				}
				last := discMessages.GetFirst()
				if last != nil {
					moreMessages, err := s.ChannelMessages(channel.ID, 100, "", last.ID, "")
					if err != nil {
						log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Error getting messages")
						return
					}
					discMessages.Push(moreMessages)
				}
			}
		}
		userMessages := &Ring{}

		if mention != nil {
			for i := 0; i < discMessages.Len(); i++ {
				message := discMessages.Get(i)
				if message.Author.ID == mention.ID {
					if cont := strings.Trim(message.Content, " "); !strings.HasPrefix(cont, viper.GetString("bot.prefix")) && len(message.Content) > 0 {
						userMessages.Push([]*discordgo.Message{message})
					}
				}
			}
			if userMessages.Len() == 0 {
				// If no messages by user, delete channel and try again
				allChannels[channelIndex] = allChannels[len(allChannels)-1]
				allChannels[len(allChannels)-1] = nil
				allChannels = allChannels[:len(allChannels)-1]
				attempts++
				continue
			}
		}
		var messages *Ring
		if userMessages.Len() > 0 {
			messages = userMessages
		} else {
			messages = discMessages
		}
		if messages.Len() == 0 {
			log.WithFields(ctx.Value(logKey).(log.Fields)).Error("Error getting messages")
			return
		}

		// quote takes into consideration nubmer of reacts for considering which message to quote, more reactions means higher chance
		weightsSlice := make([]uint, messages.Len())
		var totalWeight uint = 0
		// duplicating code inside for loop
		for i := range weightsSlice {
			uniqueReacts := messages.Get(i).Reactions
			var sumOfReacts uint = 0
			for _, uniqueReaction := range uniqueReacts {
				sumOfReacts += uint(uniqueReaction.Count)
			}
			weightsSlice[i] = sumOfReacts + 1
			totalWeight += weightsSlice[i]
		}
		randomInt := rand.Intn(int(totalWeight))
		fmt.Println(strconv.Itoa(int(totalWeight)))

		var i int = -1
		for randomInt > 0 {
			i++
			randomInt -= int(weightsSlice[i])
		}
		message := messages.Get(i)

		messageContent, err := message.ContentWithMoreMentionsReplaced(s)
		if err != nil {
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Error parsing mentions")
			return
		}
		timestamp, err := message.Timestamp.Parse()
		if err != nil {
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Error getting time")
			return
		}

		embed := NewEmbed().SetAuthor(message.Author.Username, message.Author.AvatarURL("")).SetDescription(
			fmt.Sprintf("*%s in #%s*", timestamp.Format(layoutIE), channel.Name),
		).SetTitle(messageContent).TruncateTitle()

		if len(message.Attachments) > 0 && message.Attachments[0].Width > 0 {
			embed.SetImage(message.Attachments[0].URL)
		}
		_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
		if err != nil {
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Error sending message")
		}
		return
	}
}

// dm commands
func dmCommands(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	userInput := strings.Split(m.Content, " ")[0]

	found := -1

	// ceck if dmer is registering, if not ignore messages
	for i, a := range registering {
		if a == m.Author.ID {
			found = i
			break
		}
	}

	if found == -1 {
		return
	}

	// If no verification code has been sent yet
	if _, ok := verifyCodes[m.Author.ID]; !ok {
		// Check for umail account
		if !strings.HasSuffix(userInput, "@umail.ucc.ie") {
			s.ChannelMessageSend(m.ChannelID, "Please use a valid UCC email address")
			return
		}
		rand.Seed(time.Now().UnixNano())
		// Generate phrase
		randomCode := petname.Generate(3, "-")
		// Send email
		response, err := sendEmail("server.registration@netsoc.co",
			userInput,
			"Netsoc Discord Verification",
			"Please message the following token to the Netsoc Bot to gain access to the Discord Server:\n\n"+
				randomCode+"\n\nIf you did not request access to the Netsoc Discord Server, ignore this message.")
		if err != nil {
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Failed to send email")
			s.ChannelMessageSend(m.ChannelID, "Failed to send email. Please try again later")
			return
		}
		if response.StatusCode == 200 || response.StatusCode == 202 {
			verifyCodes[m.Author.ID] = randomCode
			s.ChannelMessageSend(m.ChannelID, "Please reply with the token that has been emailed to you")
		} else {
			log.WithFields(ctx.Value(logKey).(log.Fields)).Error("Sendgrid returned status " + strconv.Itoa(response.StatusCode) + " reponse body: " + response.Body)
			s.ChannelMessageSend(m.ChannelID, "Failed to send email. Please try again later")
		}
		return
	}

	// If code sent doesnt equal verification code
	if userInput != verifyCodes[m.Author.ID] {
		s.ChannelMessageSend(m.ChannelID, "Incorrect token. Please try again")
		return
	}

	servers := viper.Get("discord.servers").(*config.Servers)
	roles := strings.Split(viper.GetString("discord.roles"), ",")

	guild, err := s.Guild(servers.PublicServer)
	if err != nil {
		log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Failed to get Public Server guild")
		return
	}

	// Add each role
	for _, roleID := range roles {
		err = s.GuildMemberRoleAdd(guild.ID, m.Author.ID, roleID)
		if err != nil {
			log.WithFields(ctx.Value(logKey).(log.Fields)).WithError(err).Error("Failed to add role " + roleID + " to user " + m.Author.ID + " in guild " + guild.ID)
			s.ChannelMessageSend(m.ChannelID, "Failed to register for the server. Please contact the owners of the server")
			return
		}
	}
	delete(verifyCodes, m.Author.ID) // Remove verify code
	// Successfully registered
	s.ChannelMessageSend(m.ChannelID, "Thank you. You have been registered for the Netsoc Discord Server")
	channels := viper.Get("discord.channels").(*config.Channels)
	s.ChannelMessageSend(channels.PublicGeneral, fmt.Sprintf("Welcome to the Netsoc Discord Server %s! Thanks for registering.", m.Author.Mention()))
	registering[found] = registering[len(registering)-1]
	registering[len(registering)-1] = ""
	registering = registering[:len(registering)-1]
}
