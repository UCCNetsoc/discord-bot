package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/Strum355/log"
	"github.com/apognu/gocal"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

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

	var events []gocal.Event
	cachedEvents, found := cached.Get("events")
	if found {
		events = cachedEvents.([]gocal.Event)
	} else {
		events = QueryCalendarEvents()
		cached.Set("events", events, cache.DefaultExpiration)
	}

	w.Header().Set("content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	returnEvents := []returnEvent{}
	for i, event := range events {
		if i == amount {
			break
		}
		eventImgURL := viper.GetString("google.calendar.image.default")
		re := regexp.MustCompile(`\/file\/d\/([^\/]+)`)
		for _, attachment := range event.Attachments {
			if attachment.Mime[:5] == "image" {
				eventImgURL = fmt.Sprintf(" https://drive.google.com/uc?id=%s", re.FindString(attachment.Value)[8:])
			}
		}
		returnEvents = append(returnEvents, returnEvent{
			event.Summary,
			event.Description,
			eventImgURL,
			event.Start,
			event.End,
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
