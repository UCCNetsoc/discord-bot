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

const publicMessageCutoff = 10

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

type sortEvents []*Event

func (e sortEvents) Len() int           { return len(e) }
func (e sortEvents) Less(i, j int) bool { return e[i].Date.Unix() < e[j].Date.Unix() }
func (e *sortEvents) Swap(i, j int) {
	temp := (*e)[i]
	(*e)[i] = (*e)[j]
	(*e)[j] = temp
}

// Run the REST API
func Run(s *discordgo.Session) {
	cached = cache.New(3*time.Minute, 3*time.Minute)
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
	sortE := sortEvents(events)
	sort.Sort(&sortE)
	if len(events) > amount {
		events = events[:amount]
	}
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	returnEvents := []returnEvent{}
	for _, event := range events {
		returnEvents = append(returnEvents, returnEvent{
			event.Title,
			event.Description,
			event.ImgURL,
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

type sortAnnouncements []*Announcement

func (a sortAnnouncements) Len() int           { return len(a) }
func (a sortAnnouncements) Less(i, j int) bool { return a[i].Date.Unix() > a[j].Date.Unix() }
func (a *sortAnnouncements) Swap(i, j int) {
	temp := (*a)[i]
	(*a)[i] = (*a)[j]
	(*a)[j] = temp
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

	var announcements []*Announcement
	cachedAnnouncements, found := cached.Get("announcements")
	if found {
		announcements = cachedAnnouncements.([]*Announcement)
	} else {
		announcements = []*Announcement{}
		privateChannelID := viper.Get("discord.channels").(*config.Channels).PrivateEvents
		privateAnnounce, err := session.ChannelMessages(privateChannelID, 100, "", "", "")
		if err != nil {
			log.WithError(err).Error("Error querying bot_events for api")
			return
		}
		publiChannelID := viper.Get("discord.channels").(*config.Channels).PublicAnnouncements
		publicAnnounce, err := session.ChannelMessages(publiChannelID, 100, "", "", "")
		if err != nil {
			log.WithError(err).Error("Error querying announcements for api")
			return
		}
		// Get messages from private command channel
		for _, event := range privateAnnounce {
			parsed, err := ParseAnnouncement(&discordgo.MessageCreate{Message: event}, "")
			if err == nil {
				// Message successfully parsed as an event.
				announcements = append(announcements, parsed)
			}
		}
		// Get other messages from public announcements.
		for _, message := range publicAnnounce {
			if message.Author.ID != session.State.User.ID && len(message.Content) > publicMessageCutoff {
				date, err := message.Timestamp.Parse()
				if err != nil {
					log.WithError(err).Error("Message time parse fail")
					return
				}
				img, err := parseImage(message)
				if err != nil {
					log.WithError(err).Error("Message image parse fail")
					return
				}
				content, err := message.ContentWithMoreMentionsReplaced(session)
				if err != nil {
					log.WithError(err).Error("Message mentions replace fail")
					return
				}
				announcement := &Announcement{
					Date:    date,
					Content: content,
					Image:   img,
				}
				announcements = append(announcements, announcement)
			}
		}
		sortA := sortAnnouncements(announcements)
		sort.Sort(&sortA)
		cached.Set("announcements", announcements, cache.DefaultExpiration)
	}
	if len(announcements) > amount {
		announcements = announcements[:amount]
	}
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	returnAnnouncements := []returnAnnouncement{}
	for _, ann := range announcements {
		announce := returnAnnouncement{
			Date:    ann.Date.Unix(),
			Content: ann.Content,
		}
		if ann.ImgData != nil {
			announce.ImageURL = ann.ImgURL
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
