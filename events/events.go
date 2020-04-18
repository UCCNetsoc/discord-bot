package events

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Strum355/log"
	"github.com/bwmarrin/discordgo"
)

// Event for use in api and bot
type Event struct {
	Title,
	Description string
	Date  time.Time
	Image *http.Response
}

const layoutISO = "2006-01-02"

// ParseEvent will give an event object
func ParseEvent(m *discordgo.MessageCreate, help string) (*Event, string) {
	// In the correct channel
	params := strings.Split(m.Content, "\"")
	if len(params) != 7 {
		return nil, fmt.Sprintf("Error parsing command\n```%s```", help)
	}
	if len(m.Attachments) != 1 || m.Attachments[0].Width == 0 {
		return nil, "No image attached"
	}
	title := params[1]
	date := params[3]
	description := params[5]
	dateTime, err := time.Parse(layoutISO, date)
	if err != nil {
		return nil, "Error parsing date. Should be in the format yyyy-mm-dd"
	}
	image := m.Attachments[0]
	imageReader, err := http.Get(image.URL)
	if err != nil {
		log.Error(err.Error())
		return nil, ""
	}

	return &Event{
		Title:       title,
		Description: description,
		Date:        dateTime,
		Image:       imageReader,
	}, ""
}
