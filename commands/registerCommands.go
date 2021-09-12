package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/prometheus"
	"github.com/bwmarrin/discordgo"
	"github.com/dghubble/oauth1"
	twitterApi "github.com/ericm/go-twitter/twitter"
	"github.com/spf13/viper"
)

var (
	// Twitter
	twitterClient *twitterApi.Client
)

// Reaction on a user message
type Reaction string

const (
	twitter Reaction = "🇹"
)

var (
	helpStrings          = make(map[string]string)
	committeeHelpStrings = make(map[string]string)
	commandsMap          = make(map[string]func(context.Context, *discordgo.Session, *discordgo.MessageCreate))
	reactionMap          = make(map[string]interface{}) // Maps message ids to content
)

type commandFunc func(context.Context, *discordgo.Session, *discordgo.MessageCreate)

func command(name string, helpMessage string, function commandFunc, committee bool) {
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
	command("version", "commit hash for the running bot version", version, false)
	command("members", "returns the number of users of the given role id", members, false)
	command("online", "see how many people are online in minecraft.netsoc.co", online, false)
	command("dig", "run a DNS query: dig TYPE DOMAIN [@RESOLVER]", digCommand, false)
	command("corona", "gives stats on corona. usage: *`!corona`* or *`!corona country-name`*", coronaCommand, false)
	command(
		"event",
		"send a message in the format: *`!event \"title\" \"yyyy-mm-dd\" \"description\"`* and make sure to have an image attached too.",
		addEvent,
		true,
	)
	command(
		"sevent",
		"same as *`!event`* but doesn't @ everyone",
		addEventSilent,
		true,
	)
	command(
		"wevent",
		"same as *`!event`* but only posts to the website, not #announcements",
		addEventWebsite,
		true,
	)
	command(
		"announce",
		"send a message in the format *`!announce TEXT`*",
		addAnnouncement,
		true,
	)
	command(
		"sannounce",
		"same as *`!announce`* but doesn't @ everyone",
		addAnnouncementSilent,
		true,
	)
	command(
		"recall",
		"PERMANENTLY DELETE the last announcement or event.",
		recall,
		true,
	)
	command(
		"upcoming",
		"replies with embed of the the next upcoming netsoc event, queried from facebook",
		upcomingEventMessage,
		false,
	)
	command("up", "check the status of various Netsoc hosted websites", checkUpCommand, true)
	command(
		"shorten",
		"shorten a URL, generating a random shortened URL if none is specified: *`!shorten original-url [shortened-slug]`* or delete a shortened url with *`!shorten delete [shortened-slug]`*",
		shortenCommand,
		true,
	)
	command("vaccines", "gives current stats on the COVID-19 vaccine rollout", vaccines, false)
	command("boosters", "check current nitro boosters", boostersCommand, false)

	// Setup APIs
	twitterConfig := oauth1.NewConfig(viper.GetString("twitter.key"), viper.GetString("twitter.secret"))
	twitterToken := oauth1.NewToken(viper.GetString("twitter.access.key"), viper.GetString("twitter.access.secret"))
	httpClient := twitterConfig.Client(oauth1.NoContext, twitterToken)
	twitterClient = twitterApi.NewClient(httpClient)

	s.AddHandler(messageCreate)
	s.AddHandler(messageReaction)
	s.AddHandler(messageDelete)
	s.AddHandler(memberLeave)
}

// Called whenever a message is sent in a server the bot has access to
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Type == discordgo.MessageTypeUserPremiumGuildSubscription && m.Author != nil {
		nitroAnnounce(s, m)
		return
	}
	if m.Author.Bot {
		return
	}
	// Check if its a DM
	if len(m.GuildID) == 0 {
		ctx := context.WithValue(context.Background(), log.Key, log.Fields{
			"author_id":  m.Author.ID,
			"channel_id": m.ChannelID,
			"guild_id":   "DM",
		})
		dmCommands(ctx, s, m)
		return
	} else {
		go prometheus.MessageCreate(m.GuildID, m.ChannelID)
	}

	if !strings.HasPrefix(m.Content, viper.GetString("bot.prefix")) {
		return
	}
	callCommand(s, m)
}

func extractCommand(c string) (commandStr string, body string) {
	body = strings.TrimPrefix(c, viper.GetString("bot.prefix"))
	commandStr = strings.Fields(body)[0]
	return
}

func callCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()
	commandStr, body := extractCommand(m.Content)
	// if command is a normal command
	if command, ok := commandsMap[commandStr]; ok {
		var channelName string
		if len(m.ChannelID) > 0 {
			if channel, err := s.Channel(m.ChannelID); err != nil {
				log.WithError(err).Error("Couldn't query channel")
				return
			} else {
				channelName = channel.Name
			}
		}
		ctx := context.WithValue(ctx, log.Key, log.Fields{
			"author_id":    m.Author.ID,
			"channel_id":   m.ChannelID,
			"guild_id":     m.GuildID,
			"user":         m.Author.Username,
			"channel_name": channelName,
			"command":      commandStr,
			"body":         body,
		})
		log.WithContext(ctx).Info("invoking standard command")
		command(ctx, s, m)
		return
	}
}

func messageReaction(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.UserID == s.State.User.ID {
		return
	}
	react := Reaction(m.MessageReaction.Emoji.Name)
	if data, ok := reactionMap[m.MessageID]; ok {
		if content, ok := data.(api.Entry); ok {
			switch react {
			case twitter:
				mediaIds := []int64{}
				image := content.GetImage()
				if image.ImgData != nil {
					// Contains image
					// Upload image
					mediaResponse, mediaHTTP, err := twitterClient.Media.Upload(&twitterApi.MediaUploadParams{
						File:     image.ImgData.Bytes(),
						MimeType: image.ImgHeader.Get("content-type"),
					})
					if err != nil {
						log.WithError(err).Error("Failed to upload image")
						return
					}
					mediaIds = append(mediaIds, mediaResponse.MediaID)
					log.Info(mediaResponse.MediaIDString)
					log.Info(mediaHTTP.Status)
				}
				// Send tweet
				tweet, _, err := twitterClient.Statuses.Update(content.GetContent(), &twitterApi.StatusUpdateParams{
					MediaIds: mediaIds,
				})
				if err != nil {
					log.WithError(err).Error("Failed to send tweet")
					return
				}
				log.Info(tweet.Text)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("https://twitter.com/%s/status/%d", tweet.User.ScreenName, tweet.ID))
			}
		}
	}
}

func memberLeave(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	prometheus.MemberJoinLeave()
}

func messageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	// Check if its not a DM
	if len(m.GuildID) != 0 {
		prometheus.MessageDelete(m.GuildID, m.ChannelID)
	}
}
