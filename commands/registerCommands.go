package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/ring"
	"github.com/bwmarrin/discordgo"
	"github.com/dghubble/oauth1"
	twitterApi "github.com/ericm/go-twitter/twitter"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

var (
	cachedMessages *cache.Cache
	globalSession  *discordgo.Session

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
	globalSession = s
	command("ping", "pong!", ping, false)
	command("help", "displays this message", help, false)
	command("register", "registers you as a member of the server", serverRegister, false)
	command("quote", "display a random quote from a netsoc member.\n\tUsage: !quote or !quote {@user,#channel}", quote, false)
	command(
		"event",
		fmt.Sprintf("send a message in the format: \n\t!event \"title\" \"yyyy-mm-dd\" \"description\" \n\tand make sure to have an image attached too.\n\tCharacter limit of %d for description", viper.GetInt("discord.charlimit")),
		addEvent,
		true,
	)
	command(
		"sevent",
		fmt.Sprintf("send a message in the format: \n\t!event \"title\" \"yyyy-mm-dd\" \"description\" \n\tand make sure to have an image attached too.\n\tCharacter limit of %d for description\n\tThis version doesn't @ everyone.", viper.GetInt("discord.charlimit")),
		addEventSilent,
		true,
	)
	command(
		"announce",
		fmt.Sprintf("send a message in the format: \n\t!announce TEXT"),
		addAnnouncement,
		true,
	)
	command(
		"sannounce",
		fmt.Sprintf("silent announce. send a message in the format: \n\t!announce TEXT\n\tThis version doesn't @ everyone."),
		addAnnouncementSilent,
		true,
	)
	command(
		"recall",
		fmt.Sprintf("PERMANENTLY DELETE the last announcement or event."),
		recall,
		true,
	)
	command("up", "check the status of various Netsoc hosted websites", checkUpCommand, true)

	// Add online command
	helpStrings["online"] = "see how many people are online in minecraft.netsoc.co"

	// Setup APIs
	twitterConfig := oauth1.NewConfig(viper.GetString("twitter.key"), viper.GetString("twitter.secret"))
	twitterToken := oauth1.NewToken(viper.GetString("twitter.access.key"), viper.GetString("twitter.access.secret"))
	httpClient := twitterConfig.Client(oauth1.NoContext, twitterToken)
	twitterClient = twitterApi.NewClient(httpClient)

	s.AddHandler(messageCreate)
	s.AddHandler(messageReaction)
	s.AddHandler(serverJoin)
}

// Called whenever a message is sent in a server the bot has access to
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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

		if !isCommittee(s, m) {
			dmCommands(ctx, s, m)
			return
		}
	}

	if !strings.HasPrefix(m.Content, viper.GetString("bot.prefix")) {
		return
	}
	callCommand(s, m)
}

func callCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()
	body := strings.TrimPrefix(m.Content, viper.GetString("bot.prefix"))

	commandStr := strings.Fields(body)[0]

	// if command is a normal command
	if command, ok := commandsMap[commandStr]; ok {
		ctx := context.WithValue(ctx, log.Key, log.Fields{
			"author_id":  m.Author.ID,
			"channel_id": m.ChannelID,
			"guild_id":   m.GuildID,
			"command":    commandStr,
			"body":       body,
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

// CacheMessages initializes message caching
func CacheMessages() {
	s := globalSession
	cachedMessages = cache.New(cache.NoExpiration, cache.NoExpiration)
	servers := viper.Get("discord.servers").(*config.Servers)

	allChannels, err := s.GuildChannels(servers.PublicServer)
	if err != nil {
		log.WithError(err).Error("Couldn't find public guild")
		return
	}

	for _, channel := range allChannels {
		if channel != nil {
			perms, err := s.UserChannelPermissions(s.State.User.ID, channel.ID)
			if err != nil {
				log.WithError(err).Error("Error getting channel perms")
				return
			}
			if channel.Type == discordgo.ChannelTypeGuildText &&
				perms&discordgo.PermissionReadMessages > 0 {
				discMessages, err := s.ChannelMessages(channel.ID, 100, "", "", "")
				ringMessages := ring.Ring{}
				ringMessages.Push(discMessages)
				if err != nil {
					log.WithError(err).Error("Error getting messages")
					return
				}
				for _j := 0; _j < 10; _j++ {
					last := ringMessages.GetLast()
					if last != nil {
						more, err := s.ChannelMessages(channel.ID, 100, last.ID, "", "")
						if err != nil {
							log.WithError(err).Error("Error getting more messages")
							return
						}
						if len(more) == 0 {
							break
						}
						ringMessages.Push(more)
					} else {
						break
					}
				}
				cachedMessages.Set(channel.ID, &ringMessages, cache.NoExpiration)
				log.Info(fmt.Sprintf("Cached %d messages for channel %s on startup", ringMessages.Len(), channel.Name))
			}
		}
	}
}
