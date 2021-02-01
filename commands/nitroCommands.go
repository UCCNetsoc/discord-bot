package commands

import (
	"context"
	"fmt"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func nitroAnnounce(s *discordgo.Session, m *discordgo.MessageCreate) {
	em := embed.NewEmbed()
	em.SetTitle("New Nitro Boost 🎉")
	em.SetDescription(fmt.Sprintf("Thank you %s for boosting the server!", m.Author.Mention()))
	em.SetColor(0xdccb01)
	channels := viper.Get("discord.channels").(*config.Channels)
	s.ChannelMessageSendEmbed(channels.PublicGeneral, em.MessageEmbed)
}

func boostersCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	servers := viper.Get("discord.servers").(*config.Servers)
	if err := s.RequestGuildMembers(servers.PublicServer, "", 0, false); err != nil {
		log.WithContext(ctx).WithError(err).Error("Couldn't query server members")
		return
	}
	guild, err := s.State.Guild(servers.PublicServer)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Couldn't query guild")
		return
	}
	boosters := []*discordgo.Member{}
	for _, member := range guild.Members {
		if len(member.PremiumSince) > 0 {
			boosters = append(boosters, member)
		}
	}
	s.GuildMembers(servers.PublicServer, "", 1000)
	em := embed.NewEmbed()
	if len(boosters) == 0 {
		s.ChannelMessageSend(m.ChannelID, "There are currently no nitro boosters for this server")
		return
	}
	desc := ""
	for i, member := range boosters {
		desc += fmt.Sprintf("\n%d\t\t**%s**", i+1, member.User.Mention())
	}
	em.SetTitle("Current Nitro Boosters")
	em.SetColor(0xdccb01)
	em.SetDescription(desc)
	s.ChannelMessageSendEmbed(m.ChannelID, em.MessageEmbed)
}
