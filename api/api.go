package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/corona"
	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

type returnEvent struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
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
	cached  *cache.Cache
	session *discordgo.Session
)

// Run the REST API
func Run(s *discordgo.Session) {
	cached = cache.New(3*time.Minute, 3*time.Minute)

	session = s

	http.HandleFunc("/events", getEvents)
	http.HandleFunc("/announcements", getAnnouncements)
	http.HandleFunc("/getMembers", getMembers)

	http.HandleFunc("/corona", postCorona)
	setWebhook()

	http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("api.port")), nil)
}

func setWebhook() {
	b := bytes.NewBuffer([]byte{})
	json.NewEncoder(b).Encode(struct {
		URL string
	}{
		viper.GetString("corona.webhook"),
	})
	resp, _ := http.Post("https://api.covid19api.com/webhook", "application/json", b)
	b.ReadFrom(resp.Body)
	log.Info(fmt.Sprintf("Corona webhook activation: %s %d", b.String(), resp.StatusCode))
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
