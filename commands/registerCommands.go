package commands

import (
	"fmt"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var (
	publicCommands = []discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "pong",
		},
		{
			Name:        "version",
			Description: "Commit hash for the running bot version",
		},
		{
			Name:        "members",
			Description: "Get number of users in the given role",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "Role option",
					Required:    true,
				},
			},
		},
		{
			Name:        "dig",
			Description: "Run a DNS query",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "type",
					Description: "Record type",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "A",
							Value: 0,
						},
						{
							Name:  "NS",
							Value: 1,
						},
						{
							Name:  "CNAME",
							Value: 2,
						},
						{
							Name:  "SRV",
							Value: 3,
						},
						{
							Name:  "TXT",
							Value: 4,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "domain",
					Description: "Domain name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "resolver",
					Description: "Query resolver",
					Required:    false,
				},
			},
		},
		{
			Name:        "corona",
			Description: "Gives current stats on corona case numbers",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "country",
					Description: "Query by country",
					Required:    false,
				},
			},
		},
		{
			Name:        "vaccines",
			Description: "Gives current stats on the COVID-19 vaccine rollout",
		},
		{
			Name:        "boosters",
			Description: "Check current nitro boosters",
		},
		{
			Name:        "upcoming",
			Description: "Gives an embed of the the next upcoming netsoc event",
		},
		{
			Name:        "online",
			Description: "See how many people are online in minecraft.netsoc.co",
		},
	}

	committeeCommands = []discordgo.ApplicationCommand{
		{
			Name:        "up",
			Description: "Check the status of various Netsoc hosted websites",
		},
		{
			Name:        "shorten",
			Description: "URL Shortener interactions",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "create",
					Description: "Create a shortened URL, the shortened URL is random if none is specified",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "original-url",
							Description: "Original URL",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "shortened-slug",
							Description: "Shortened Slug",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "delete",
					Description: "Delete a shortened URL",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "shortened-slug",
							Description: "Shortened Slug",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List all shortened URL's",
				},
			},
		},
	}
)

func RegisterCommands(s *discordgo.Session) {
	/* TODO: Edit permissions for public commands, currently not supported by discordGo, might need to manually send a bulk edit request
	(https://discord.com/developers/docs/interactions/application-commands#batch-edit-application-command-permissions)*/
	for _, command := range publicCommands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, viper.GetString("discord.public.server"), &command)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Cannot create slash command %q: %v", command.Name, err))
		}
		_, err = s.ApplicationCommandCreate(s.State.User.ID, viper.GetString("discord.committee.server"), &command)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Cannot create slash command %q: %v", command.Name, err))
		}
	}
	for _, command := range committeeCommands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, viper.GetString("discord.committee.server"), &command)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Cannot create slash command %q: %v", command.Name, err))
		}
	}
}
