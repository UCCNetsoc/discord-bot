package api

import "net/http"

// Entry represents an entry in the announcements channel
type Entry interface {
	GetContent() string
	GetImage() *http.Response
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
func (a Announcement) GetImage() *http.Response {
	return a.Image
}

// GetImage returns Image response data
func (e Event) GetImage() *http.Response {
	return e.Image
}
