package config

import "github.com/spf13/viper"

func initDefaults() {
	// Discord
	viper.SetDefault("discord.token", "") // GitHub scrapers be like -.-
	viper.SetDefault("discord.servers", &Servers{})
	// Consul
	viper.SetDefault("consul.address", "consul:8500")
	viper.SetDefault("consul.token", "")
}
