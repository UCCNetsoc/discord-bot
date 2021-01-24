package utils

import (
	"encoding/json"
	"errors"

	"github.com/bwmarrin/discordgo"
)

// A GuildPreview holds data related to a specific public Discord Guild, even if the user is not in the guild.
type GuildPreview struct {
	// The ID of the guild.
	ID string `json:"id"`

	// The name of the guild. (2–100 characters)
	Name string `json:"name"`

	// The hash of the guild's icon. Use Session.GuildIcon
	// to retrieve the icon itself.
	Icon string `json:"icon"`

	// The hash of the guild's splash.
	Splash string `json:"splash"`

	// The hash of the guild's discovery splash.
	DiscoverySplash string `json:"discovery_splash"`

	// The list of enabled guild features
	Features []string `json:"features"`

	// Approximate number of members in this guild, returned from the GET /guild/<id> endpoint when with_counts is true
	ApproximateMemberCount int `json:"approximate_member_count"`

	// Approximate number of non-offline members in this guild, returned from the GET /guild/<id> endpoint when with_counts is true
	ApproximatePresenceCount int `json:"approximate_presence_count"`

	// the description for the guild
	Description string `json:"description"`
}

func unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return errors.New("Unmarshalling error")
	}

	return nil
}

// GuildPreview returns a GuildPreview structure of a specific public Guild.
// guildID   : The ID of a Guild
func GetGuildPreview(s *discordgo.Session, guildID string) (st *GuildPreview, err error) {
	body, err := s.RequestWithBucketID("GET", discordgo.EndpointGuild(guildID)+"/preview", nil, discordgo.EndpointGuild(guildID)+"/preview")
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}
