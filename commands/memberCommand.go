package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Strum355/log"

	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func members(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(strings.TrimPrefix(m.Content, viper.GetString("bot.prefix")+"members"))

	servers := viper.Get("discord.servers").(*config.Servers)
	// publicServer, err := s.Guild(servers.PublicServer)

	var num int
	role, err := s.State.Role(servers.PublicServer, args[0])
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("failed to get role by ID")
		s.ChannelMessageSend(m.ChannelID, "Failed to find role by that ID")
		return
	}

	members, err := s.GuildMembers(servers.PublicServer, "", 1000)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("failed to get members")
		return
	}

	if role.Name == "@everyone" {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Number of members is %d", len(members)))
		return
	}

	for _, member := range members {
		for _, roleID := range member.Roles {
			if role.ID == roleID {
				num++
			}
		}
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Number of members in %s is %d", role.Name, num))
}
