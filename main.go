package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/UCCNetsoc/discord-bot/commands"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var production bool

func main() {
	// Check for flags
	production = *flag.Bool("p", false, "enables production with json logging")
	flag.Parse()
	if production {
		log.InitJSONLogger(&log.Config{Output: os.Stdout})
	} else {
		log.InitSimpleLogger(&log.Config{Output: os.Stdout})
	}

	// Setup viper and consul
	exitError(config.InitConfig())

	// Discord connection
	token := viper.GetString("discord.token")
	session, err := discordgo.New("Bot " + token)
	exitError(err)
	// Open websocket
	err = session.Open()
	commands.Register(session)
	exitError(err)

	// Maintain connection until a SIGTERM, then cleanly exit
	log.Info("Bot is Running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Info("Cleanly exiting")
	session.Close()
}

func exitError(err error) {
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
