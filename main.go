package main

import (
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
	production = *flag.Bool("p", false, "enables production with json logging")
	flag.Parse()
	if production {
		log.InitJSONLogger(&log.Config{Output: os.Stdout})
	} else {
		log.InitSimpleLogger(&log.Config{Output: os.Stdout})
	}

	// Consul config
	consulConfig := api.DefaultConfig()
	if production {
		consulConfig.Scheme = "https"
	} else {
		consulConfig.Scheme = "http"
	}
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
	comitteeServer, _, err := kv.Get("COMMITTEE_SERVER", nil)
	if err != nil {
		log.Error(err.Error())
		return
	}
	config.CommitteeServer = string(comitteeServer.Value)

	// Get Public Server
	publicServer, _, err := kv.Get("PUBLIC_SERVER", nil)
	if err != nil {
		log.Error(err.Error())
		return
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
