package main

import (
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
	PublicServer,
	CommitteeServer string
}

var (
	config     *Config
	production bool
)

func main() {
	config = &Config{}
	// Check for flags
	for _, flag := range os.Args {
		if flag == "-p" {
			production = true
		}
	}
	if production {
		log.InitJSONLogger(&log.Config{Output: os.Stdout})
	} else {
		log.InitSimpleLogger(&log.Config{Output: os.Stdout})
	}

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

func readConfig() {
	// Connect to consul
	client, err := api.NewClient(api.DefaultConfig())
	exitError(err)
	kv := client.KV()

	// Get Commmittee Server
	comitteeServer, _, err := kv.Get("COMMITTEE_SERVER", nil)
	if err != nil {
		log.Error(err.Error())
	}
	config.CommitteeServer = string(comitteeServer.Value)

	// Get Public Server
	publicServer, _, err := kv.Get("PUBLIC_SERVER", nil)
	if err != nil {
		log.Error(err.Error())
	}
	config.PublicServer = string(publicServer.Value)

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
