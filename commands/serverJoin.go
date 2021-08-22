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
			s.ChannelMessageSendEmbed(welcomeID, emb.MessageEmbed)
		}
	}
}
