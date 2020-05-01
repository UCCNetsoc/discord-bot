package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Strum355/log"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

var cached *cache.Cache

// Run the REST API
func Run() {
	cached = cache.New(30*time.Minute, time.Hour)

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
	}
	w.Header().Set("content-type", "application/json")
	b, err := json.Marshal(events)
	if err != nil {
		log.WithFields(log.Fields{"events": events}).WithError(err).Error("Error marshalling events")
	}
	w.Write(b)
}

func getAnnouncements(w http.ResponseWriter, r *http.Request) {

}
