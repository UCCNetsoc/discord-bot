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
			log.Error(err.Error())
		}
		log.WithFields(log.Fields{"value": string(structData.Value)}).Info("Consul updated")

	}
	go watch.Run(viper.GetString("consul.address"))
	return nil
}