package config

import (
	"encoding/json"
	"fmt"

	"github.com/Strum355/log"
	"github.com/hashicorp/consul/api"
	consulwatch "github.com/hashicorp/consul/api/watch"
	"github.com/spf13/viper"
)

// Servers represents the servers pulled from Consul
type Servers struct {
	PublicServer    string `json:"public"`
	CommitteeServer string `json:"committee"`
}

// Channels required for events
type Channels struct {
	PublicAnnouncements string `json:"public_announcements"` // On public server
	PublicGeneral       string `json:"public_general"`       // On public server
	PrivateEvents       string `json:"private_events"`       // On committee server
}

// ReadFromConsul sets up watchers for consul
func ReadFromConsul(callback func()) error {
	// Connect to consul
	// Listen for consul updates

	err := createConsulWatcher("servers", "discord.servers", callback)
	if err != nil {
		return fmt.Errorf("error creating watcher for 'discord.servers': %w", err)
	}

	err = createConsulWatcher("welcome_messages", "discord.welcome_messages")
	if err != nil {
		return fmt.Errorf("error creating watcher for 'discord.welcome_messages': %w", err)
	}

	err = createConsulWatcher("channels", "discord.channels")
	if err != nil {
		return fmt.Errorf("error creating watcher for 'discord.channels': %w", err)
	}

	err = createConsulWatcher("quote_blacklist", "discord.quote_blacklist")
	if err != nil {
		return fmt.Errorf("error creating watcher for 'discord.quote_blacklist': %w", err)
	}

	return nil
}

func createConsulWatcher(consulKey, viperKey string, callbacks ...func()) error {
	consulKey = "discordbot/" + consulKey
	// Channels watcher
	params := map[string]interface{}{
		"type":  "key",
		"key":   consulKey,
		"token": viper.GetString("consul.token"),
	}
	watch, err := consulwatch.Parse(params)
	if err != nil {
		return err
	}

	watch.Handler = func(idx uint64, data interface{}) {
		channels := viper.Get(viperKey)
		structData, ok := data.(*api.KVPair)
		if !ok {
			log.WithFields(log.Fields{
				"type":       fmt.Sprintf("%T", data),
				"viper_key":  viperKey,
				"consul_key": consulKey,
			}).Error("KV malformed")
			return
		}

		if len(structData.Value) == 0 {
			log.Error("servers key doesnt exist")
			return
		}
		err = json.Unmarshal(structData.Value, channels)
		if err != nil {
			log.WithFields(log.Fields{
				"data":       string(structData.Value),
				"type":       fmt.Sprintf("%T", viper.Get(viperKey)),
				"viper_key":  viperKey,
				"consul_key": consulKey,
			}).WithError(err).Error("Failed to unmarshal Consul data")
		}
		log.WithFields(log.Fields{"value": string(structData.Value)}).Info("Consul updated")
		for _, callback := range callbacks {
			callback()
		}

	}
	go func() {
		err := watch.Run(viper.GetString("consul.address"))
		if err != nil {
			log.WithFields(log.Fields{
				"viper_key":  viperKey,
				"consul_key": consulKey,
			}).WithError(err).Error("consul watcher error")
		}
	}()
	return nil
}
