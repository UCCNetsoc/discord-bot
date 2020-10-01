package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

// Entry represents an entry in the announcements channel
type Entry interface {
	GetContent() string
	GetImage() *Image
}

// Image to be embedded in an entry
type Image struct {
	ImgData   *bytes.Buffer
	ImgHeader *http.Header
	ImgURL    string
}

// GetContent returns message content
func (a Announcement) GetContent() string {
	return a.Content
}

// GetContent returns message content
func (e Event) GetContent() string {
	return e.Description
}

// GetImage returns Image response data
func (a Announcement) GetImage() *Image {
	return a.Image
}

// GetImage returns Image response data
func (e Event) GetImage() *Image {
	return e.Image
}

func parseImage(m *discordgo.Message) (*Image, error) {
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
		defer image.Body.Close()
	}
	return &Image{
		ImgData:   imageBody,
		ImgURL:    imageURL,
		ImgHeader: imageHeader,
	}, nil
}
