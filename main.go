package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/consul/api"
	consulwatch "github.com/hashicorp/consul/api/watch"
)

// Config represents the config pulled from Consul
type Config struct {
	PublicServer    string `json:"public"`
	CommitteeServer string `json:"committee"`
}

var (
	config     *Config
	production bool
)

func main() {
	config = &Config{}
	// Check for flags
	production = *flag.Bool("p", false, "enables production with json logging")
	flag.Parse()
	if production {
		log.InitJSONLogger(&log.Config{Output: os.Stdout})
	} else {
		log.InitSimpleLogger(&log.Config{Output: os.Stdout})
	}

	// Consul config
	consulConfig := api.DefaultConfig()
	var err error
	consulConfig.Address, err = getEnv("CONSUL_ADDRESS")
	exitError(err)

	readConfig(consulConfig)

	// Discord connection
	token, err := getEnv("DISCORD_TOKEN")
	exitError(err)
	session, err := discordgo.New("Bot " + token)
	exitError(err)
	// Open websocket
	err = session.Open()
	exitError(err)

	// Maintain connection until a SIGTERM, then cleanly exit
	log.Info("Bot is Running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Info("Cleanly exiting")
	session.Close()
}

func readConfig(consulConfig *api.Config) {
	var err error
	// Connect to consul
	// Listen for consul updates
	params := map[string]interface{}{
		"type": "key",
		"key":  "discordbot/servers",
	}
	params["token"] = os.Getenv("CONSUL_TOKEN")
	watch, err := consulwatch.Parse(params)
	exitError(err)
	watch.Handler = func(idx uint64, data interface{}) {
		structData, ok := data.(*api.KVPair)
		if !ok {
			log.Error("KV malformed")
			return
		}
		if len(structData.Value) == 0 {
			log.Error("servers key doesnt exist")
			return
		}
		err = json.Unmarshal(structData.Value, config)
		if err != nil {
			log.Error(err.Error())
		}
		log.WithFields(log.Fields{"value": string(structData.Value)}).Info("Consul updated")

	}
	go watch.Run(consulConfig.Address)
}

func getEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("Error finding %s", key)
	}
	return val, nil
}

func exitError(err error) {
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
