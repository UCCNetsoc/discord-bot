package commands

import (
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
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
)

var registering = make([]string, 0)
var verifyCodes = make(map[string]string)

const layoutIE = "02/01/06"

// ping command
func ping(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "pong")
	if err != nil {
		log.WithError(err).Error("Failed to send pong message")
		return
	}
}

// help command
func help(s *discordgo.Session, m *discordgo.MessageCreate) {
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
		log.WithError(err).Error("Failed to send help message")
		return
	}
}

// register command
func serverRegister(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, a := range registering {
		if a == m.Author.ID {
			return
		}
	}
	registering = append(registering, m.Author.ID)

	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		log.WithError(err).Error("Failed to create DM channel")
		return
	}

	s.ChannelMessageSend(channel.ID, "Please message me your UCC email address so I can verify you as a member of UCC")
}

func serverJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	servers := viper.Get("discord.servers").(*config.Servers)
	publicServer, err := s.Guild(servers.PublicServer)
	if err != nil {
		log.WithError(err).Error("Failed to get Public Server guild")
		return
	}
	if m.GuildID != publicServer.ID {
		return
	}
	// Handle join messages
	messages := viper.Get("discord.welcomemessages").([]string)
	if len(messages) > 0 {
		i := rand.Intn(len(messages))
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			log.WithError(err).Error("Couldnt find guild for welcome")
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
		serverRegister(s, &discordgo.MessageCreate{Message: &discordgo.Message{Author: m.User}})
		return
	}
}

func addEvent(s *discordgo.Session, m *discordgo.MessageCreate) {
	channels := viper.Get("discord.channels").(*config.Channels)
	if isCommittee(m) && m.ChannelID == channels.PrivateEvents {
		event, err := api.ParseEvent(m, committeeHelpStrings["event"])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
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

func addAnnouncement(s *discordgo.Session, m *discordgo.MessageCreate) {
	channels := viper.Get("discord.channels").(*config.Channels)
	if isCommittee(m) && m.ChannelID == channels.PrivateEvents {
		announcement, err := api.ParseAnnouncement(m, committeeHelpStrings["announce"])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		if announcement.Image != nil {
			s.ChannelFileSendWithMessage(
				channels.PublicAnnouncements,
				fmt.Sprintf("@everyone\n%s", announcement.Content),
				"poster.jpg",
				announcement.Image.Body,
			)
		} else {
			s.ChannelMessageSend(channels.PublicAnnouncements, fmt.Sprintf("@everyone\n%s", announcement.Content))
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "This command is unavailable")
	}
}

// recall events and announcements
func recall(s *discordgo.Session, m *discordgo.MessageCreate) {
	channels := viper.Get("discord.channels").(*config.Channels)
	if isCommittee(m) && m.ChannelID == channels.PrivateEvents {
		public, err := s.ChannelMessages(channels.PublicAnnouncements, 100, "", "", "")
		if err != nil {
			log.WithError(err).Error("Error getting channel public")
		}
		private, err := s.ChannelMessages(channels.PrivateEvents, 100, "", "", "")
		if err != nil {
			log.WithError(err).Error("Error getting channel private")
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
			} else if strings.HasPrefix(message.Content, viper.GetString("bot.prefix")+"event"+" ") {
				create := &discordgo.MessageCreate{Message: message}
				event, err := api.ParseEvent(create, committeeHelpStrings["event"])
				if err == nil {
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
}

// dm commands
func dmCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
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
			log.WithError(err).Error("Failed to send email")
			s.ChannelMessageSend(m.ChannelID, "Failed to send email. Please try again later")
			return
		}
		if response.StatusCode == 200 || response.StatusCode == 202 {
			verifyCodes[m.Author.ID] = randomCode
			s.ChannelMessageSend(m.ChannelID, "Please reply with the token that has been emailed to you")
		} else {
			log.Error("Sendgrid returned status " + strconv.Itoa(response.StatusCode) + " reponse body: " + response.Body)
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
		log.WithError(err).Error("Failed to get Public Server guild")
		return
	}

	for _, member := range guild.Members {
		// If member is in public server
		if member.User.ID == m.Author.ID {
			// Add each role
			for _, roleID := range roles {
				err = s.GuildMemberRoleAdd(guild.ID, m.Author.ID, roleID)
				if err != nil {
					log.WithError(err).Error("Failed to add role " + roleID + " to user " + m.Author.ID + " in guild " + guild.ID)
					s.ChannelMessageSend(m.ChannelID, "Failed to register for the server. Please contact the owners of the server")
					return
				}
			}
		}
		break
	}
	// Successfully registered
	s.ChannelMessageSend(m.ChannelID, "Thank you. You have been registered for the Netsoc Discord Server")
	registering[found] = registering[len(registering)-1]
	registering[len(registering)-1] = ""
	registering = registering[:len(registering)-1]
}

func sendEmail(from string, to string, subject string, content string) (*rest.Response, error) {
	fromAddress := mail.NewEmail(from, from)
	toAddress := mail.NewEmail(to, to)
	message := mail.NewSingleEmail(fromAddress, subject, toAddress, content, content)
	client := sendgrid.NewSendClient(viper.GetString("sendgrid.token"))
	response, err := client.Send(message)
	return response, err
}

func isCommittee(m *discordgo.MessageCreate) bool {
	return m.GuildID == (viper.Get("discord.servers").(*config.Servers).CommitteeServer)
}
