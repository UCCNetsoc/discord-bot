package commands

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/bwmarrin/discordgo"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/spf13/viper"
)

var (
	umailRegex  = regexp.MustCompile("[0-9]{8,11}@umail.ucc.ie")
	registering = make(map[string]registeringState)
	verifyCodes = make(map[string]string)
)

// registeringState defines a state node in the registering flow FSM
type registeringState func(context.Context, *discordgo.Session, *discordgo.MessageCreate) registeringState

func initiatedRegistration(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) registeringState {
	content := strings.TrimSpace(m.Content)

	if !umailRegex.MatchString(content) {
		s.ChannelMessageSend(m.ChannelID, "Please use a valid UCC email address")
		return initiatedRegistration
	}

	rand.Seed(time.Now().UnixNano())
	randomCode := petname.Generate(3, "-")
	response, err := sendEmail(
		"discord.registration@netsoc.co",
		content,
		"Netsoc Discord Verification",
		"Please message the following token to the Netsoc Bot to gain access to the Discord Server:\n\n"+
			randomCode+"\n\nIf you did not request access to the Netsoc Discord Server, ignore this message.")
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error("failed to send verification email")
		s.ChannelMessageSend(m.ChannelID, "Failed to send verification email. Please try again later or contact a SysAdmin")
		return initiatedRegistration
	}

	// Success
	if response.StatusCode < 300 && response.StatusCode > 199 {
		verifyCodes[m.Author.ID] = randomCode
		s.ChannelMessageSend(m.ChannelID, "Please reply with the token that has been emailed to you")
		return submittedEmail
	}

	log.WithContext(ctx).
		WithFields(log.Fields{"status_code": response.StatusCode, "response": response.Body}).
		Error("Sendgrid returned bad status code")
	s.ChannelMessageSend(m.ChannelID, "Failed to send verification email. Please try again later or contact a SysAdmin")
	return initiatedRegistration
}

// submittedEmail state is entered when the user has supplied a valid email address and the email was successfully
// sent. This state loops back to itself until the user supplies the correct token that is stored
func submittedEmail(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) registeringState {
	content := strings.TrimSpace(m.Content)
	code, ok := verifyCodes[m.Author.ID]
	if !ok {
		// if we're here, shits either no bueno..or the bot was restarted since
		log.WithContext(ctx).Error("expected verification token but none was found")
		s.ChannelMessageSend(m.ChannelID, "There was an issue verifying you, please contact a SysAdmin :(")
		return submittedEmail
	}

	if content != code {
		log.WithContext(ctx).
			WithFields(log.Fields{"expected_code": code, "received_code": content}).
			Warn("user supplied non-matching verification token")

		s.ChannelMessageSend(m.ChannelID, "Incorrect token. Please try again or contact a SysAdmin")
		return submittedEmail
	}

	servers := viper.Get("discord.servers").(*config.Servers)
	roles := strings.Split(viper.GetString("discord.roles"), ",")

	for _, roleID := range roles {
		err := s.GuildMemberRoleAdd(servers.PublicServer, m.Author.ID, roleID)
		if err != nil {
			log.WithContext(ctx).
				WithError(err).
				WithFields(log.Fields{"role_id": roleID, "target_guild_id": servers.PublicServer}).
				Error("failed to add role to user")
			s.ChannelMessageSend(m.ChannelID, "Failed registering you for the server, please contact a SysAdmin :(")
			return submittedEmail
		}
	}

	s.ChannelMessageSend(m.ChannelID, "Congrats! You've been registered for the Netsoc Discord Server. Have fun!")
	channels := viper.Get("discord.channels").(*config.Channels)
	s.ChannelMessageSend(channels.PublicGeneral, fmt.Sprintf("Welcome to the Netsoc Discord Server %s! Thanks for registering.", m.Author.Mention()))
	delete(registering, m.Author.ID)
	delete(verifyCodes, m.Author.ID)
	return nil
}
