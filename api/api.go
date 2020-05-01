package api

import (
	"fmt"
	"net/http"

	"github.com/spf13/viper"
)

// Run the REST API
func Run() {
	http.HandleFunc("/events", getEvents)
	http.HandleFunc("/announcements", getAnnouncements)
	http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("api.port")), nil)
}

func getEvents(w http.ResponseWriter, r *http.Request) {

}

func getAnnouncements(w http.ResponseWriter, r *http.Request) {

}
