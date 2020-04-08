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
	// Connect to consul
	client, err := api.NewClient(consulConfig)
	exitError(err)
	kv := client.KV()

	// Get Commmittee Server
	servers, _, err := kv.Get("servers", nil)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if servers != nil {
		err = json.Unmarshal(servers.Value, config)
		if err != nil {
			log.Error("Consul servers malformed entry")
		}
		log.WithFields(log.Fields{
			"key":   servers.Key,
			"value": string(servers.Value),
		}).Info("Found KV pair in consul")
	} else {
		log.Error("No Consul entry for servers")
	}

	log.Info("Successfully read config from consul")
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
