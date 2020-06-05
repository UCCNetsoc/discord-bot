package config

import "github.com/spf13/viper"

func initDefaults() {
	// Bot
	viper.SetDefault("bot.prefix", "!")
	viper.SetDefault("bot.quote.default_message_weight", 1)
	// Discord
	viper.SetDefault("discord.token", "") // GitHub scrapers be like -.-
	viper.SetDefault("discord.servers", &Servers{})
	viper.SetDefault("discord.welcome_messages", &[]string{})
	viper.SetDefault("discord.roles", "")
	viper.SetDefault("discord.autoregister", true)
	viper.SetDefault("discord.channels", &Channels{})
	viper.SetDefault("discord.charlimit", 280) // Limit for event description
	viper.SetDefault("discord.quote_blacklist", &[]string{})
	// Consul
	viper.SetDefault("consul.address", "consul:8500")
	viper.SetDefault("consul.token", "")
	// Sendgrid
	viper.SetDefault("sendgrid.token", "")
	// Twitter
	viper.SetDefault("twitter.key", "")
	viper.SetDefault("twitter.secret", "")
	viper.SetDefault("twitter.access.key", "")
	viper.SetDefault("twitter.access.secret", "")
	// Rest API
	viper.SetDefault("api.port", 80)
	viper.SetDefault("api.event_query_limit", 20)
	viper.SetDefault("api.announcement_query_limit", 20)
}
