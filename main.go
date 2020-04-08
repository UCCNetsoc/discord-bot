package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
)

func getEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("Error finding %s", key)
	}
	return val, nil
}

func main() {
	log.InitSimpleLogger(&log.Config{Output: os.Stdout})

	token, err := getEnv("DISCORD_TOKEN")
	if err != nil {
		panic(err)
	}
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		panic("Couldn't init discord connection")
	}
	// Open websocket
	err = session.Open()
	if err != nil {
		panic(err)
	}

	// Maintain connection until a SIGTERM, then cleanly exit
	log.Info("Bot is Running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Info("Cleanly exiting")
	session.Close()
}
