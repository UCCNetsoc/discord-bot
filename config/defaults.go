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
	viper.SetDefault("discord.public.corona", "")
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
	// Google Calendar
	viper.SetDefault("google.calendar.ics", "")
	// Rest API
	viper.SetDefault("api.port", 80)
	viper.SetDefault("api.event_query_limit", 20)
	viper.SetDefault("api.announcement_query_limit", 20)
	viper.SetDefault("api.public_message_cutoff", 10)
	viper.SetDefault("api.remove_symbols", []string{"@everyone", "@here"})
	// Up sites
	viper.SetDefault("netsoc.sites", "https://uccexpress.ie,http://netsoc.co,https://motley.ie,https://hlm.netsoc.co,https://uccnetsoc.netsoc.co,https://wiki.netsoc.co")
	viper.SetDefault("minecraft.host", "minecraft.netsoc.co:1194")
	// Prometheus exporter
	viper.SetDefault("prom.port", 2112)
	viper.SetDefault("prom.dbname", "promexporter")
	// Database
	viper.SetDefault("sql.host", "postgres.netsoc.local")
	viper.SetDefault("sql.port", 5432)
	viper.SetDefault("sql.username", "root")
	viper.SetDefault("sql.password", "password")
	// Corona
	viper.SetDefault("corona.default", "ireland")
	viper.SetDefault("corona.webhook", "https://events.netsoc.dev/corona")
	viper.SetDefault("freeimage.key", "6d207e02198a847aa98d0a2a901485a5")

	viper.SetDefault("shorten.domain", "links.netsoc.co")
	viper.SetDefault("shorten.username", "")
	viper.SetDefault("shorten.password", "")
}
