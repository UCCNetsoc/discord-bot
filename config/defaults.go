package config

import "github.com/spf13/viper"

func initDefaults() {
	// Bot
	viper.SetDefault("bot.prefix", "!")
	viper.SetDefault("bot.quote.default_message_weight", 1)
	viper.SetDefault("bot.version", "development")
	// Discord
	viper.SetDefault("discord.token", "") // GitHub scrapers be like -.-
	viper.SetDefault("discord.servers", &Servers{})
	viper.SetDefault("discord.channels", &Channels{})

	viper.SetDefault("discord.public.server", "")
	viper.SetDefault("discord.public.channel", "")
	viper.SetDefault("discord.public.general", "")
	viper.SetDefault("discord.public.welcome", "")
	viper.SetDefault("discord.committee.server", "")
	viper.SetDefault("discord.committee.channel", "")

	viper.SetDefault("discord.welcome_messages", &[]string{})

	viper.SetDefault("discord.roles", "")
	viper.SetDefault("discord.autoregister", true)
	viper.SetDefault("discord.charlimit", 280) // Limit for event description
	viper.SetDefault("discord.quote_blacklist", &[]string{})
	// Sendgrid
	viper.SetDefault("sendgrid.token", "")
	// Twitter
	viper.SetDefault("twitter.key", "")
	viper.SetDefault("twitter.secret", "")
	viper.SetDefault("twitter.access.key", "")
	viper.SetDefault("twitter.access.secret", "")
	// Facebook
	viper.SetDefault("facebook.appID", "1809168605905513")
	viper.SetDefault("facebook.app.secret", "1507576e6ccdcf31fc8af828c856420f")
	viper.SetDefault("facebook.pageID", "106033244624746")
	viper.SetDefault("facebook.page.access.token", "EAAZAtbeQYAmkBANnZCK4MYJvvXLJvBQSI6DentkTiT3NIx50A4vI8tnYKcpHX9PgOmBcG6hAw80JP5pSsrsdO7cAwgDbgsAuMI2cX8g1Qo2okUEpIcXcnSxHQPGqx0mWspMwOL4rPThwqB8geVD42AeCZCgJE0oOTZB9hAurZCKBWGe1vZAZAUJbUcIkvbmbZBEZD")
	// Rest API
	viper.SetDefault("api.port", 80)
	viper.SetDefault("api.event_query_limit", 20)
	viper.SetDefault("api.announcement_query_limit", 20)
	viper.SetDefault("api.public_message_cutoff", 10)
	viper.SetDefault("api.remove_symbols", []string{"@everyone", "@here"})
	// Up sites
	viper.SetDefault("netsoc.sites", "https://uccexpress.ie,https://netsoc.co,https://motley.ie,https://admin.netsoc.co,https://hlm.netsoc.co,https://uccnetsoc.netsoc.co,https://wiki.netsoc.co")
	viper.SetDefault("minecraft.host", "games.vm.netsoc.co:1194")
	// Prometheus exporter
	viper.SetDefault("prom.port", 2112)
	viper.SetDefault("prom.dbname", "promexporter")
	// Database
	viper.SetDefault("mysql.url", "mysql.netsoc.local:3306")
	viper.SetDefault("mysql.username", "root")
	viper.SetDefault("mysql.password", "password")
}
