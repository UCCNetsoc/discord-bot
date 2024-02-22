package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

// Announcement for bot and rest api
type Announcement struct {
	Date    time.Time
	Content string
	*Image
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
		publiChannelID := viper.Get("discord.channels").(*config.Channels).PublicAnnouncements
		publicAnnounce, err := session.ChannelMessages(publiChannelID, 100, "", "", "")
		if err != nil {
			log.WithError(err).Error("Error querying announcements for api")
			return
		}
		// Get other messages from public announcements.
		for _, message := range publicAnnounce {
			if message.Author.ID != session.State.User.ID && len(message.Content) > viper.GetInt("api.public_message_cutoff") {
				date, err := time.Parse(time.RFC3339, message.Timestamp.String())
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
