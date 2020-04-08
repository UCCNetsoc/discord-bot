package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
)

var production bool

func main() {
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
