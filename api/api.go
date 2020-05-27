package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

type returnEvent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Date        int64  `json:"date"`
}
type returnAnnouncement struct {
	Date     int64  `json:"date"`
	Content  string `json:"content"`
	ImageURL string `json:"image_url"`
}

var (
	cached  *cache.Cache
	session *discordgo.Session
)

type sortEvents struct {
	events []*Event
}

func (e sortEvents) Len() int           { return len(e.events) }
func (e sortEvents) Less(i, j int) bool { return e.events[i].Date.Unix() < e.events[j].Date.Unix() }
func (e *sortEvents) Swap(i, j int) {
	temp := e.events[i]
	e.events[i] = e.events[j]
	e.events[j] = temp
}

// Run the REST API
func Run(s *discordgo.Session) {
	cached = cache.New(5*time.Minute, time.Hour)
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
	if err != nil {
		http.Error(w, "Please provide an int as 'q's value", 403)
		return
	}
	if amount > limit {
		http.Error(w, "Query amount exceeds the query limit", 403)
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
		for _, event := range liveEvents {
			parsed, err := ParseEvent(&discordgo.MessageCreate{Message: event}, "")
			if err == nil {
				// Message successfully parsed as an event.
				events = append(events, parsed)
			}
		}
		cached.Set("events", events, cache.DefaultExpiration)
	}
	// Filter out events that have already passed
	allEvents := make([]*Event, len(events))
	copy(allEvents, events)
	events = []*Event{}
	for _, event := range allEvents {
		if event.Date.Unix() > time.Now().AddDate(0, 0, 1).Unix() {
			events = append(events, event)
		}
	}
	sort.Sort(&sortEvents{events})
	if len(events) > amount {
		events = events[:amount]
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
	if err != nil {
		http.Error(w, "Please provide an int as 'q's value", 403)
		return
	}
	if amount > limit {
		http.Error(w, "Query amount exceeds the query limit", 403)
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
		for _, event := range liveAnnounce {
			parsed, err := ParseAnnouncement(&discordgo.MessageCreate{Message: event}, "")
			if err == nil {
				// Message successfully parsed as an event.
				announcemenents = append(announcemenents, parsed)
			}
		}
		cached.Set("announcements", announcemenents, cache.DefaultExpiration)
	}
	if len(announcemenents) > amount {
		announcemenents = announcemenents[:amount]
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
