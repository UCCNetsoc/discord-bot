package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	session.Close()
}
