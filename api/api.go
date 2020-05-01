package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

type returnEvent struct {
	Title,
	Description string
	ImageURL string `json:"image_url"`
	Date     int64
}
type returnAnnouncement struct {
	Date     int64
	Content  string
	ImageURL string `json:"image_url"`
}

var (
	cached  *cache.Cache
	session *discordgo.Session
)

// Run the REST API
func Run(s *discordgo.Session) {
	cached = cache.New(30*time.Minute, time.Hour)
	session = s

	http.HandleFunc("/events", getEvents)
	http.HandleFunc("/announcements", getAnnouncements)

	http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("api.port")), nil)
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit := viper.GetInt("api.event_query_limit")
	queryAmount, exists := query["q"]
	if !exists || len(query) == 0 {
		http.Error(w, "Please add the parameter 'q'", 403)
		return
	}
	amount, err := strconv.Atoi(queryAmount[0])
	if err != nil || amount > limit {
		http.Error(w, "Please provide an int as 'q's value", 403)
		return
	}
	var events []*Event
	cachedEvents, found := cached.Get("events")
	if found {
		events = cachedEvents.([]*Event)
	} else {
		events = []*Event{}
		channelID := viper.Get("discord.channels").(*config.Channels).PrivateEvents
		liveEvents, err := session.ChannelMessages(channelID, 100, "", "", "")
		if err != nil {
			log.WithError(err).Error("Error querying events for api")
			return
		}
		i := 0
		for _, event := range liveEvents {
			if i == amount {
				break
			}
			parsed, err := ParseEvent(&discordgo.MessageCreate{Message: event}, "")
			if err == nil {
				// Message successfully parsed as an event.
				events = append(events, parsed)
				i++
			}
		}
		cached.Set("events", events, cache.DefaultExpiration)
	}
	w.Header().Set("content-type", "application/json")
	returnEvents := []returnEvent{}
	for _, event := range events {
		returnEvents = append(returnEvents, returnEvent{
			event.Title,
			event.Description,
			event.Image.Request.URL.String(),
			event.Date.Unix(),
		})
	}

	b, err := json.Marshal(returnEvents)
	if err != nil {
		log.WithFields(log.Fields{"events": returnEvents}).WithError(err).Error("Error marshalling events")
		return
	}
	w.Write(b)
}

func getAnnouncements(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit := viper.GetInt("api.announcement_query_limit")
	queryAmount, exists := query["q"]
	if !exists || len(query) == 0 {
		http.Error(w, "Please add the parameter 'q'", 403)
		return
	}
	amount, err := strconv.Atoi(queryAmount[0])
	if err != nil || amount > limit {
		http.Error(w, "Please provide an int as 'q's value", 403)
		return
	}
	var announcemenents []*Announcement
	cachedAnnouncements, found := cached.Get("announcements")
	if found {
		announcemenents = cachedAnnouncements.([]*Announcement)
	} else {
		announcemenents = []*Announcement{}
		channelID := viper.Get("discord.channels").(*config.Channels).PrivateEvents
		liveAnnounce, err := session.ChannelMessages(channelID, 100, "", "", "")
		if err != nil {
			log.WithError(err).Error("Error querying events for api")
			return
		}
		i := 0
		for _, event := range liveAnnounce {
			if i == amount {
				break
			}
			parsed, err := ParseAnnouncement(&discordgo.MessageCreate{Message: event}, "")
			if err == nil {
				// Message successfully parsed as an event.
				announcemenents = append(announcemenents, parsed)
				i++
			}
		}
		cached.Set("announcements", announcemenents, cache.DefaultExpiration)
	}
	w.Header().Set("content-type", "application/json")
	returnAnnouncements := []returnAnnouncement{}
	for _, ann := range announcemenents {
		announce := returnAnnouncement{
			Date:    ann.Date.Unix(),
			Content: ann.Content,
		}
		if ann.Image != nil {
			announce.ImageURL = ann.Image.Request.URL.String()
		}
		returnAnnouncements = append(returnAnnouncements, announce)
	}

	b, err := json.Marshal(returnAnnouncements)
	if err != nil {
		log.WithFields(log.Fields{"announcements": returnAnnouncements}).WithError(err).Error("Error marshalling announcements")
		return
	}
	w.Write(b)

}
