package commands

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/UCCNetsoc/discord-bot/embed"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/emails"
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

// initiatedRegistration state is entered when a user invokes the register command to join the server. This state
// loops back to itself until the user supplies a valid umail email and the verification email sends successfully
func initiatedRegistration(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) registeringState {
	content := strings.TrimSpace(m.Content)
	log.WithContext(ctx).Info("Emailing user.")

	if !umailRegex.MatchString(content) {
		s.ChannelMessageSendEmbed(m.ChannelID, errorEmbed("Please use a valid UCC email address!"))
		return initiatedRegistration
	}

	rand.Seed(time.Now().UnixNano())
	randomCode := petname.Generate(3, "-")
	response, err := emails.SendEmail(
		"discord.registration@netsoc.co",
		content,
		"UCC Netsoc Discord Verification",
		"Please message the following token to the Netsoc Bot to gain access to the UCC Netsoc Discord Server:\n\n"+
			randomCode+"\n\nIf you did not request access to the UCC Netsoc Discord Server, ignore this message.",
		emails.FillTemplate(
			"Discord Verification",
			"Please message the following token to the Netsoc Bot to gain access to the UCC Netosc Discord Server. <br /><br />If you did not request access to the UCC Netsoc Discord Server, ignore this message.",
			randomCode),
	)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error("failed to send verification email")
		s.ChannelMessageSendEmbed(m.ChannelID, errorEmbed("Failed to send verification email. Please try again later or contact a SysAdmin"))
		return initiatedRegistration
	}

	// Success
	if response.StatusCode < 300 && response.StatusCode > 199 {
		verifyCodes[m.Author.ID] = randomCode
		s.ChannelMessageSendEmbed(m.ChannelID, embed.NewEmbed().
			SetTitle("UCC Netsoc Server Registration").
			SetDescription("Please reply with the token that has been emailed to you.").
			MessageEmbed)
		return submittedEmail
	}

	log.WithContext(ctx).
		WithFields(log.Fields{"status_code": response.StatusCode, "response": response.Body}).
		Error("Sendgrid returned bad status code")
	s.ChannelMessageSendEmbed(m.ChannelID, errorEmbed("Failed to send verification email. Please try again later or contact a SysAdmin"))
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
		s.ChannelMessageSendEmbed(m.ChannelID, errorEmbed("There was an issue verifying you, please contact a SysAdmin :("))
		return submittedEmail
	}

	if content != code {
		log.WithContext(ctx).
			WithFields(log.Fields{"expected_code": code, "received_code": content}).
			Warn("user supplied non-matching verification token")

		s.ChannelMessageSendEmbed(m.ChannelID, errorEmbed("Incorrect token. Please try again or contact a SysAdmin"))
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
			s.ChannelMessageSendEmbed(m.ChannelID, errorEmbed("Failed registering you for the server, please contact a SysAdmin :("))
			return submittedEmail
		}
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed.NewEmbed().SetTitle("✔️ Verified!").SetDescription("Congrats! You've been registered for the Netsoc Discord Server. Have fun!").MessageEmbed)
	channels := viper.Get("discord.channels").(*config.Channels)

	s.ChannelMessageSendEmbed(channels.PublicGeneral, embed.NewEmbed().
		SetTitle("Welcome").
		SetDescription(fmt.Sprintf("Welcome to the Netsoc Discord Server %s! Thanks for registering.", m.Author.Mention())).
		MessageEmbed)

	delete(registering, m.Author.ID)
	delete(verifyCodes, m.Author.ID)
	return nil
}
