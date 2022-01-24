package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
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
		events, err = QueryCalendarEvents(viper.GetString("google.calendar.public.ics"))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
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
		if len(event.Attachments) > 0 {
			for _, attachment := range event.Attachments {
				if attachment.Mime[:5] == "image" {
					if strings.Contains(attachment.Value, "drive.google.com/file/d/") {
						id := strings.Split(attachment.Value, "/d/")[1]
						id = strings.Split(id, "/view")[0]
						eventImgURL = "https://drive.google.com/uc?export=download&id=" + id
					} else if strings.Contains(attachment.Value, "drive.google.com/open?id=") {
						id := strings.Split(attachment.Value, "open?id=")[1]
						eventImgURL = "https://drive.google.com/uc?export=download&id=" + id
					}
				}
			}
		}
		formattedDescription := strings.ReplaceAll(event.Description, `\n`, "\n")
		returnEvents = append(returnEvents, returnEvent{
			event.Summary,
			formattedDescription,
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

func QueryCalendarEvents(url string) ([]gocal.Event, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.WithError(err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err)
		return nil, err
	}

	r := bytes.NewBuffer(body)

	start, end := time.Now(), time.Now().Add(30*24*time.Hour)

	c := gocal.NewParser(r)
	c.Start, c.End = &start, &end
	err = c.Parse()
	if err != nil {
		return nil, err
	}

	calEvents := c.Events
	sort.SliceStable(calEvents, func(i, j int) bool {
		return calEvents[i].Start.Before(*calEvents[j].Start)
	})

	return calEvents, nil
}
