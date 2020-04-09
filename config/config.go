package config

import (
	"strings"

	"github.com/spf13/viper"
)

// InitConfig sets up viper and consul
func InitConfig() error {
	// Viper
	initDefaults()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // For gamers only
	viper.AutomaticEnv()

	// Consul
	return readFromConsul()
}
