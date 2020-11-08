package commands

import (
	"context"
	"time"

	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/embed"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"github.com/spf13/viper"
)

var (
	committeeMembers []*discordgo.Member
	lastUpdated      time.Time
)

func isCommittee(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if m.GuildID != "" {
		return m.GuildID == (viper.Get("discord.servers").(*config.Servers).CommitteeServer)
	}

	var err error
	// Check if committee list has been updated in the last 5 minutes
	if time.Now().Sub(lastUpdated) >= (5 * time.Minute) {
		committeeMembers, err = s.GuildMembers(viper.Get("discord.servers").(*config.Servers).CommitteeServer, "", 100)
		if err != nil {
			log.WithFields(log.Fields{
				"guildID": viper.Get("discord.servers").(*config.Servers).CommitteeServer,
			}).WithError(err).Error("Could not fetch committee members")
			return false
		}
		lastUpdated = time.Now()
	}

	return memberContains(committeeMembers, m.Author.ID)
}

// Helper function to check for a user ID in a list of members
func memberContains(members []*discordgo.Member, userID string) bool {
	found := false
	for _, member := range members {
		if member.User.ID == userID {
			found = true
			break
		}
	}
	return found
}

func errorEmbed(message string) *discordgo.MessageEmbed {
	return embed.NewEmbed().SetTitle("❗ ERROR ❗").SetDescription(message).MessageEmbed
}

func RefeshSchedule(schedule *gocron.Scheduler, s *discordgo.Session) {
	schedule.Clear()
	upcomingEvents, err := api.QueryFacebookEvents()
	if err != nil {
		log.WithError(err).Error("Error occured refrshing cron scheduler event")
	}
	if len(upcomingEvents) > 0 {
		schedule.Every(1).Hour().StartAt((time.Unix((upcomingEvents[0].Date - 600), 0))).Do(UpcomingEventAnnounce, context.TODO(), s)
	}
}
