package commands

import (
	"math/rand"
	"strings"
	"time"

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

// ping command
func ping(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "pong")
	if err != nil {
		log.Error(err.Error())
		return
	}
}

// help command
func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	out := "```"
	for k, v := range helpStrings {
		out += k + ": " + v + "\n"
	}
	_, err := s.ChannelMessageSend(m.ChannelID, out+"```")
	if err != nil {
		log.Error(err.Error())
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
		log.Error(err.Error())
		return
	}

	s.ChannelMessageSend(channel.ID, "Please message me your UCC email address so I can verify you as a member of UCC")
}

// dm commands
func dmCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	word := strings.Split(m.Content, " ")[0]

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
		if !strings.HasSuffix(word, "@umail.ucc.ie") {
			s.ChannelMessageSend(m.ChannelID, "Please use a valid UCC email address")
			return
		}
		rand.Seed(time.Now().UnixNano())
		// Generate phrase
		randomCode := petname.Generate(3, "-")
		// Send email
		response, err := sendEmail("server.registration@netsoc.co",
			word,
			"Netsoc Discord Verification",
			"Please message the following token to the Netsoc Bot to gain access to the Discord Server:\n\n"+
				randomCode+"\n\nIf you did not request access to the Netsoc Discord Server, ignore this message.")
		if err != nil {
			log.Error(err.Error())
			s.ChannelMessageSend(m.ChannelID, "Failed to send email. Please try again later")
			return
		}
		if response.StatusCode == 200 || response.StatusCode == 202 {
			verifyCodes[m.Author.ID] = randomCode
			s.ChannelMessageSend(m.ChannelID, "Please reply with the token that has been emailed to you")
		} else {
			log.Error(response.Body)
			s.ChannelMessageSend(m.ChannelID, "Failed to send email. Please try again later")
		}
		return
	}

	// If code sent doesnt equal verification code
	if word != verifyCodes[m.Author.ID] {
		s.ChannelMessageSend(m.ChannelID, "Incorrect token. Please try again")
		return
	}

	servers := viper.Get("discord.servers").(*config.Servers)
	roles := strings.Split(viper.GetString("discord.roles"), ",")

	guild, err := s.Guild(servers.PublicServer)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, member := range guild.Members {
		// If member is in public server
		if member.User.ID == m.Author.ID {
			// Add each role
			for _, roleID := range roles {
				err = s.GuildMemberRoleAdd(guild.ID, m.Author.ID, roleID)
				if err != nil {
					log.Error(err.Error())
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
