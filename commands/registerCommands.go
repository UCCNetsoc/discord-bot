package commands

import (
	"context"
	"fmt"

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
	commandsMap = make(map[string]func(context.Context, *discordgo.Session, *discordgo.InteractionCreate))
	reactionMap = make(map[string]interface{}) // Maps message ids to content
)

type commandFunc func(context.Context, *discordgo.Session, *discordgo.InteractionCreate)

func command(name string, function commandFunc) {
	commandsMap[name] = function
}

// Register commands
func Register(s *discordgo.Session) {
	// Public commands
	command("ping", ping)
	command("version", version)
	command("members", members)
	command("dig", dig)
	command("corona", coronaCommand)
	command("vaccines", vaccines)
	command("boosters", boostersCommand)
	command("upcoming", upcomingEvent)
	command("online", online)
	// Committee commands
	command("up", checkUpCommand)
	command("shorten", shortenCommand)

	// Setup APIs
	twitterConfig := oauth1.NewConfig(viper.GetString("twitter.key"), viper.GetString("twitter.secret"))
	twitterToken := oauth1.NewToken(viper.GetString("twitter.access.key"), viper.GetString("twitter.access.secret"))
	httpClient := twitterConfig.Client(oauth1.NoContext, twitterToken)
	twitterClient = twitterApi.NewClient(httpClient)

	// Setup Interaction Handlers
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		callCommand(s, i)
	})

	// Setup Message Handlers
	s.AddHandler(messageCreate)
	s.AddHandler(messageReaction)
	s.AddHandler(messageDelete)
	s.AddHandler(serverJoin)
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
	if len(m.GuildID) != 0 {
		go prometheus.MessageCreate(m.GuildID, m.ChannelID)
	}
}

// Returns useful data about the command's contents
func extractCommandContent(i *discordgo.InteractionCreate) (commandAuthor *discordgo.User, commandName string, commandBody []string) {
	commandAuthor = i.Member.User
	// Location of the data changes with regard to the type of interaction
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		commandName = i.ApplicationCommandData().Name
		if len(i.ApplicationCommandData().Options) > 0 {
			for _, option := range i.ApplicationCommandData().Options {
				commandBody = append(commandBody, fmt.Sprintf("%s : %v", option.Name, option.Value))
			}
		}
	case discordgo.InteractionMessageComponent:
		commandName = i.MessageComponentData().CustomID
		for idx, value := range i.MessageComponentData().Values {
			commandBody = append(commandBody, fmt.Sprintf("value %d : %s", idx, value))
		}
	}
	return
}

func callCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		log.WithError(err).Error("Couldn't query channel")
		return
	}

	commandAuthor, commandName, commandBody := extractCommandContent(i)
	if command, ok := commandsMap[commandName]; ok {
		ctx := context.WithValue(ctx, log.Key, log.Fields{
			"author_id":    commandAuthor.ID,
			"channel_id":   i.ChannelID,
			"guild_id":     i.GuildID,
			"user":         commandAuthor.Username,
			"channel_name": channel.Name,
			"command":      commandName,
			"body":         commandBody,
		})

		log.WithContext(ctx).Info("invoking standard command")
		command(ctx, s, i)
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
