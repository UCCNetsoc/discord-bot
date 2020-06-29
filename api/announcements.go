package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	var (
		image       *http.Response
		err         error
		imageHeader *http.Header
		imageURL    string
		imageBody   *bytes.Buffer
	)
	if len(m.Attachments) > 0 && m.Attachments[0].Width > 0 {
		image, err = http.Get(m.Attachments[0].URL)
		if err != nil {
			return nil, fmt.Errorf("Error parsing image: %w", err)
		}
		imageHeader = &image.Header
		imageURL = image.Request.URL.String()
		imageRead, err := ioutil.ReadAll(image.Body)
		if err != nil {
			return nil, err
		}
		imageBody = bytes.NewBuffer(imageRead)
		fmt.Println(imageBody.Len(), image)
		defer image.Body.Close()
	}
	date, err := m.Timestamp.Parse()
	if err != nil {
		return nil, fmt.Errorf("Error coverting date: %w", err)
	}
	return &Announcement{
		Date:    date,
		Content: content,
		Image: &Image{
			ImgData:   imageBody,
			ImgURL:    imageURL,
			ImgHeader: imageHeader,
		},
	}, nil
}
