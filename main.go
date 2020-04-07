package main

import (
	"fmt"
	"os"

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
	session, err := discordgo.New("")
	if err != nil {
		panic("Couldn't init discord connection")
	}
}
