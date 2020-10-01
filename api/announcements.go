package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

// Announcement for bot and rest api
type Announcement struct {
	Date    time.Time
	Content string
	*Image
}

// ParseAnnouncement Return an annoucement from a message
func ParseAnnouncement(m *discordgo.MessageCreate, help string) (*Announcement, error) {
	// In the correct channel
	var content string
	if strings.HasPrefix(m.Content, viper.GetString("bot.prefix")+"announce") {
		content = strings.TrimPrefix(m.Content, viper.GetString("bot.prefix")+"announce")
		content = strings.Trim(content, " ")
	} else if strings.HasPrefix(m.Content, viper.GetString("bot.prefix")+"sannounce") {
		content = strings.TrimPrefix(m.Content, viper.GetString("bot.prefix")+"sannounce")
		content = strings.Trim(content, " ")
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("Error parsing command\n```%s```", help)
	}
	date, err := m.Timestamp.Parse()
	if err != nil {
		return nil, fmt.Errorf("Error coverting date: %w", err)
	}
	img, err := parseImage(m.Message)
	if err != nil {
		return nil, err
	}
	return &Announcement{
		Date:    date,
		Content: content,
		Image:   img,
	}, nil
}
