package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Strum355/log"
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
func ParseAnnouncement(m *discordgo.MessageCreate, help string) (*Announcement, string) {
	// In the correct channel
	content := strings.TrimPrefix(m.Content, viper.GetString("bot.prefix")+"announce")
	if len(content) == 0 {
		return nil, fmt.Sprintf("Error parsing command\n```%s```", help)
	}
	var image *http.Response
	var err error
	if len(m.Attachments) > 0 && m.Attachments[0].Width > 0 {
		image, err = http.Get(m.Attachments[0].URL)
		if err != nil {
			log.Error(err.Error())
			return nil, ""
		}
	}
	date, err := m.Timestamp.Parse()
	if err != nil {
		log.Error(err.Error())
		return nil, ""
	}
	return &Announcement{
		date,
		content,
		image,
	}, ""
}
