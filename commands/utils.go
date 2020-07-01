package commands

import (
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func isCommittee(m *discordgo.MessageCreate) bool {
	return m.GuildID == (viper.Get("discord.servers").(*config.Servers).CommitteeServer)
}
