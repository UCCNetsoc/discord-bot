package main

import (
	"encoding/base64"
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
	// Connect to consul
	// Listen for consul updates
	params := map[string]interface{}{
		"type": "key",
		"key":  "discordbot/servers",
	}
	watch, err := consulwatch.Parse(params)
	exitError(err)
	watch.Handler = func(idx uint64, data interface{}) {
		// Serialize data
		buf, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			log.Error(err.Error())
		}
		// Get it's Value
		structData := &struct{ Value string }{}
		err = json.Unmarshal(buf, structData)
		if err != nil {
			log.Error(err.Error())
		}
		if len(structData.Value) == 0 {
			log.Error("servers key doesnt exist")
			return
		}
		// Decode base64 Value
		decoded, err := base64.RawStdEncoding.DecodeString(
			structData.Value[:len(structData.Value)-1],
		)
		if err != nil {
			log.Error(err.Error())
		}
		err = json.Unmarshal(decoded, config)
		if err != nil {
			log.Error(err.Error())
		}
		log.WithFields(log.Fields{"response": string(buf), "value": string(decoded)}).Info("Consul updated")

	}
	go watch.Run(consulConfig.Address)

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
