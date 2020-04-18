package commands

import (
	"fmt"
	"strings"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var helpStrings = make(map[string]string)
var committeeHelpStrings = make(map[string]string)
var commandsMap = make(map[string]func(*discordgo.Session, *discordgo.MessageCreate))

func command(name string, helpMessage string, function func(*discordgo.Session, *discordgo.MessageCreate), committee bool) {
	if committee {
		committeeHelpStrings[name] = helpMessage
	} else {
		helpStrings[name] = helpMessage
	}
	commandsMap[name] = function
}

// Register commands
func Register(s *discordgo.Session) {
	command("ping", "pong!", ping, false)
	command("help", "displays this message", help, false)
	command("register", "registers you as a member of the server", serverRegister, false)
	command(
		"event",
		fmt.Sprintf("send a message in the format: \n\t!event \"title\" \"mm/dd/yyyy\" \"description\" \n\tand make sure to have an image attached too.\n\tCharacter limit of %d for description", viper.GetInt("discord.charlimit")),
		addEvent,
		true,
	)
	command(
		"announce",
		fmt.Sprintf("send a message in the format: \n\t!announce TEXT\n\tCharacter limit of %d", viper.GetInt("discord.charlimit")),
		addAnnouncement,
		true,
	)

	s.AddHandler(messageCreate)
	s.AddHandler(serverJoin)
}

// Called whenever a message is sent in a server the bot has access to
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	for k, v := range commandsMap {
		if !(m.Content == viper.GetString("bot.prefix")+k || strings.HasPrefix(m.Content, viper.GetString("bot.prefix")+k+" ")) {
			continue
		}
		v(s, m)
	}

	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		log.WithError(err).Error("Failed to create DM channel")
		return
	}
	// DM
	if channel.ID == m.ChannelID {
		dmCommands(s, m)
	}
}
