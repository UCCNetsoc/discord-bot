package commands

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

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
		SetTitle("UCC Netsoc Server Registration").
		SetDescription("Send me your UCC email address so we can verify you're a UCC student.").
		SetFooter("Message your email in the form <Student ID>@umail.ucc.ie. A code will be sent to your email that you will then send here.")
	s.ChannelMessageSendEmbed(channel.ID, emb.MessageEmbed)
}

func serverJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	ctx := context.WithValue(context.Background(), log.Key, log.Fields{
		"user_id":  m.User.ID,
		"guild_id": m.GuildID,
	})
	log.WithContext(ctx).Info(fmt.Sprintf("%s has joined the server", m.Member.User.Username))
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
				emb.SetFooter("We've sent you a DM so you can register for full access to the server!\nIf you're a student in another college simply let us know here and we will be able to assign you a role manually!\n*If you don't receive the email right away, remember it can take up to 5 minutes to go through. If you still haven' received it, check your spam folder ;)*")
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
