package main

import (
	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session

func main() {
	// TODO: Get credentials from vault
	var err error
	session, err = discordgo.New("")
	if err != nil {
		panic("Couldn't init discord connection")
	}
}
