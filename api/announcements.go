package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

// Announcement for bot and rest api
type Announcement struct {
	Date    time.Time
	Content string
	Image   *http.Response
}

// ParseAnnouncement Return an annoucement from a message
func ParseAnnouncement(m *discordgo.MessageCreate, help string) (*Announcement, error) {
	// In the correct channel
	content := strings.TrimPrefix(m.Content, viper.GetString("bot.prefix")+"announce")
	content = strings.Trim(content, " ")
	if len(content) == 0 {
		return nil, fmt.Errorf("Error parsing command\n```%s```", help)
	}
	limit := viper.GetInt("discord.charlimit")
	if len(content) > limit {
		return nil, fmt.Errorf("Announcement exceeds %d characters", limit)
	}
	var image *http.Response
	var err error
	if len(m.Attachments) > 0 && m.Attachments[0].Width > 0 {
		image, err = http.Get(m.Attachments[0].URL)
		if err != nil {
			return nil, fmt.Errorf("Error parsing image: %w", err)
		}
	}
	date, err := m.Timestamp.Parse()
	if err != nil {
		return nil, fmt.Errorf("Error coverting date: %w", err)
	}
	return &Announcement{
		date,
		content,
		image,
	}, nil
}
