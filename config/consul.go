package config

import (
	"encoding/json"

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
	PrivateEvents       string `json:"private_events"`       // On committee server
}

func readFromConsul() error {
	var err error
	// Connect to consul
	// Listen for consul updates

	// Servers
	params := map[string]interface{}{
		"type":  "key",
		"key":   "discordbot/servers",
		"token": viper.GetString("consul.token"),
	}
	watch, err := consulwatch.Parse(params)
	if err != nil {
		return err
	}
	watch.Handler = func(idx uint64, data interface{}) {
		servers := viper.Get("discord.servers").(*Servers)
		structData, ok := data.(*api.KVPair)
		if !ok {
			log.Error("KV malformed")
			return
		}
		if len(structData.Value) == 0 {
			log.Error("servers key doesnt exist")
			return
		}
		err = json.Unmarshal(structData.Value, servers)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal " + string(structData.Value) + " to Servers struct")
		}
		log.WithFields(log.Fields{"value": string(structData.Value)}).Info("Consul updated")

	}
	go watch.Run(viper.GetString("consul.address"))

	// Welcome Messages
	paramsWelcome := map[string]interface{}{
		"type":  "key",
		"key":   "discordbot/welcomemessages",
		"token": viper.GetString("consul.token"),
	}
	watchWelcome, err := consulwatch.Parse(paramsWelcome)
	if err != nil {
		return err
	}
	watchWelcome.Handler = func(idx uint64, data interface{}) {
		messages := viper.Get("discord.welcomemessages").([]string)
		structData, ok := data.(*api.KVPair)
		if !ok {
			log.Error("KV malformed")
			return
		}
		if len(structData.Value) == 0 {
			log.Error("servers key doesnt exist")
			return
		}
		err = json.Unmarshal(structData.Value, &messages)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal " + string(structData.Value) + " to Servers struct")
		}
		log.WithFields(log.Fields{"value": string(structData.Value)}).Info("Consul updated")

	}
	go watchWelcome.Run(viper.GetString("consul.address"))

	// Channels watcher
	paramsChannels := map[string]interface{}{
		"type":  "key",
		"key":   "discordbot/channels",
		"token": viper.GetString("consul.token"),
	}
	watchChannels, err := consulwatch.Parse(paramsChannels)
	if err != nil {
		return err
	}
	watchChannels.Handler = func(idx uint64, data interface{}) {
		channels := viper.Get("discord.channels").(*Channels)
		structData, ok := data.(*api.KVPair)
		if !ok {
			log.Error("KV malformed")
			return
		}
		if len(structData.Value) == 0 {
			log.Error("servers key doesnt exist")
			return
		}
		err = json.Unmarshal(structData.Value, channels)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal " + string(structData.Value) + " to Servers struct")
		}
		log.WithFields(log.Fields{"value": string(structData.Value)}).Info("Consul updated")

	}
	go watchChannels.Run(viper.GetString("consul.address"))
	return nil
}
