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

// Event for use in api and bot
type Event struct {
	Title,
	Description string
	Date time.Time
	*Image
}

const layoutISO = "2006-01-02"

// ParseEvent will give an event object
func ParseEvent(m *discordgo.MessageCreate, help string) (*Event, error) {
	// In the correct channel
	params := strings.Split(m.Content, "\"")
	if len(params) != 7 {
		return nil, fmt.Errorf("Error parsing command\n```%s```", help)
	}
	if len(m.Attachments) != 1 || m.Attachments[0].Width == 0 {
		return nil, fmt.Errorf("No image attached")
	}
	title := params[1]
	date := params[3]
	description := params[5]
	limit := viper.GetInt("discord.charlimit")
	if len(description) > limit {
		return nil, fmt.Errorf("Description exceeds %d characters", limit)
	}
	dateTime, err := time.Parse(layoutISO, date)
	if err != nil {
		return nil, fmt.Errorf("Error parsing date. Should be in the format yyyy-mm-dd")
	}
	image := m.Attachments[0]
	imageReader, err := http.Get(image.URL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing image: %w", err)
	}
	defer imageReader.Body.Close()
	imageRead, err := ioutil.ReadAll(imageReader.Body)
	if err != nil {
		return nil, err
	}
	imageBody := bytes.NewBuffer(imageRead)

	return &Event{
		Title:       title,
		Description: description,
		Date:        dateTime,
		Image: &Image{
			ImgData:   imageBody,
			ImgURL:    imageReader.Request.URL.String(),
			ImgHeader: &imageReader.Header,
		},
	}, nil
}
