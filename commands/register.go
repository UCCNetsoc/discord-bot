package commands

import (
	"strings"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var helpStrings = make(map[string]string)
var commandsMap = make(map[string]func(*discordgo.Session, *discordgo.MessageCreate))

func command(name string, helpMessage string, function func(*discordgo.Session, *discordgo.MessageCreate)) {
	helpStrings[name] = helpMessage
	commandsMap[name] = function
}

//Register commands
func Register(s *discordgo.Session) {
	command("ping", "pong!", ping)
	command("help", "displays this message", help)
	command("register", "registers you as a member of the server", serverRegister)

	s.AddHandler(messageCreate)
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
		log.Error(err.Error())
		return
	}
	// DM
	if channel.ID == m.ChannelID {
		dmCommands(s, m)
	}
}
