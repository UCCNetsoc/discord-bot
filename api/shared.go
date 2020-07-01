package api

import (
	"bytes"
	"net/http"
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
