package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/corona"
	"github.com/apognu/gocal"
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

type returnCalendarEvent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

type returnAnnouncement struct {
	Date     int64  `json:"date"`
	Content  string `json:"content"`
	ImageURL string `json:"image_url"`
}
type returnMembers struct {
	Count int `json:"count"`
}

var (
	cached    *cache.Cache
	session   *discordgo.Session
	cachedCal *cache.Cache
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
	cachedCal = cache.New(1*time.Minute, 1*time.Minute)

	session = s

	http.HandleFunc("/events", getEvents)
	http.HandleFunc("/announcements", getAnnouncements)
	http.HandleFunc("/getMembers", getMembers)
	http.HandleFunc("/calendarEvents", getCalendarEvents)

	http.HandleFunc("/corona", postCorona)
	setWebook()

	http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("api.port")), nil)
}

func setWebook() {
	b := bytes.NewBuffer([]byte{})
	json.NewEncoder(b).Encode(struct {
		URL string
	}{
		viper.GetString("corona.webhook"),
	})
	resp, _ := http.Post("https://api.covid19api.com/webhook", "application/json", b)
	b.ReadFrom(resp.Body)
	log.Info(fmt.Sprintf("Corona webhook actiavtion: %s %d", b.String(), resp.StatusCode))
}

func postCorona(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	total := &corona.TotalSummary{}
	if err := json.NewDecoder(r.Body).Decode(total); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	country := total.GetCountry(viper.GetString("corona.default"))
	log.WithContext(r.Context()).Info("New COVID data. Sending.")
	embs, err := corona.CreateEmbed(country, session, r.Context())
	if err != nil {
		return
	}
	for _, emb := range embs {
		session.ChannelMessageSendEmbed(viper.GetString("discord.public.corona"), emb)
	}
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit := viper.GetInt("api.event_query_limit")
	queryAmount, exists := query["q"]
	if !exists || len(query) == 0 {
		http.Error(w, "Please add the parameter 'q'", 400)
		return
	}
	amount, err := strconv.Atoi(queryAmount[0])
	if err != nil {
		http.Error(w, "Please provide an int as 'q's value", 400)
		return
	}
	if amount > limit {
		http.Error(w, "Query amount exceeds the query limit", 400)
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
		if event.Date.Unix() > time.Now().AddDate(0, 0, 0).Unix() {
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

func getCalendarEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var events []gocal.Event
	cachedEvents, found := cachedCal.Get("events")
	if found {
		events = cachedEvents.([]gocal.Event)
	} else {
		events = QueryCalendarEvents()
		cached.Set("events", events, cache.DefaultExpiration)
	}

	returnEvents := []returnCalendarEvent{}
	for _, event := range events {
		returnEvents = append(returnEvents, returnCalendarEvent{
			event.Summary,
			event.Description,
			// event.Attachments,
			event.Start.String(),
			event.End.String(),
		})
	}
	b, err := json.Marshal(returnEvents)
	if err != nil {
		log.WithFields(log.Fields{"events": returnEvents}).WithError(err).Error("Error marshalling events")
		return
	}
	w.Write(b)

}

func QueryCalendarEvents() []gocal.Event {
	resp, err := http.Get(viper.GetString("google.calendar.ics"))
	if err != nil {
		log.WithError(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err)
	}

	r := bytes.NewBuffer(body)

	start, end := time.Now(), time.Now().Add(30*24*time.Hour)

	c := gocal.NewParser(r)
	c.Start, c.End = &start, &end
	c.Parse()

	return c.Events
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
		http.Error(w, "Please add the parameter 'q'", 400)
		return
	}
	amount, err := strconv.Atoi(queryAmount[0])
	if err != nil {
		http.Error(w, "Please provide an int as 'q's value", 400)
		return
	}
	if amount > limit {
		http.Error(w, "Query amount exceeds the query limit", 400)
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
			if message.Author.ID != session.State.User.ID && len(message.Content) > viper.GetInt("api.public_message_cutoff") {
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
				var replaced bool
				for _, symbol := range viper.GetStringSlice("api.remove_symbols") {
					if strings.Contains(content, symbol) {
						replaced = true
						content = strings.ReplaceAll(content, symbol, "")
					}
				}
				if !replaced {
					continue
				}
				content = strings.TrimSpace(content)
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

func getMembers(w http.ResponseWriter, r *http.Request) {
	servers := viper.Get("discord.servers").(*config.Servers)
	members, err := session.GuildMembers(servers.PublicServer, "", 1000)
	if err != nil {
		log.WithError(err).Error("Failed to get members")
		http.Error(w, "Failed to get members", 500)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	json.NewEncoder(w).Encode(returnMembers{Count: len(members)})
}
