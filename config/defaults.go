package config

import "github.com/spf13/viper"

func initDefaults() {
	// Bot
	viper.SetDefault("bot.prefix", "!")
	// Discord
	viper.SetDefault("discord.token", "") // GitHub scrapers be like -.-
	viper.SetDefault("discord.servers", &Servers{})
	viper.SetDefault("discord.welcomemessages", []string{})
	viper.SetDefault("discord.roles", "")
	viper.SetDefault("discord.autoregister", true)
	viper.SetDefault("discord.channels", &Channels{})
	viper.SetDefault("discord.charlimit", 280) // Limit for event description
	// Consul
	viper.SetDefault("consul.address", "consul:8500")
	viper.SetDefault("consul.token", "")
	// Sendgrid
	viper.SetDefault("sendgrid.token", "")
}
