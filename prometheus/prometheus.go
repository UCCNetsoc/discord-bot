package prometheus

import (
	"fmt"
	"net/http"
	"strings"

	"database/sql"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/bwmarrin/discordgo"

	// Needed for mysql
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

var (
	memberCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "member_count",
		Help: "The total number of members in the server",
	})
	membersJoined = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "members_joined",
		Help: "The total number of members to have ever joined the server",
	})
	eventCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "event_count",
		Help: "The total number of events ran",
	})
	messageCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "message_count",
		Help: "The total number of messages sent in the server",
	},
		[]string{
			"server",
			"channel",
		})
	globalSession *discordgo.Session
	globalDB      *sql.DB
)

// MemberJoin is called whenever a member joins the server
// Increments memberCount and increments membersJoined if member hasn't joined in the past
func MemberJoin(id string) {
	result, err := globalDB.Query("SELECT id FROM joined WHERE id = " + id)
	defer result.Close()
	if err != nil {
		log.WithError(err).Error("Failed to get joined")
		return
	}
	if !result.Next() {
		membersJoined.Inc()
		_, err = globalDB.Exec("INSERT INTO joined VALUES(" + id + ")")
		if err != nil {
			log.WithError(err).Error("Failed to add id to joined")
			return
		}
	}

	memberCount.Inc()

}

// MemberLeave is called whenever a member leaves the server
// Recalculated memberCount based on number of members with roles
// This is due to not being able to determine whether a leaving member had a role
func MemberLeave(id string) {
	count := 0.
	servers := viper.Get("discord.servers").(*config.Servers)
	publicServer, err := globalSession.Guild(servers.PublicServer)
	if err != nil {
		log.WithError(err).Error("Failed to get Public Server guild")
		return
	}
	for _, member := range publicServer.Members {
		found := false
		for _, roleID := range strings.Split(viper.GetString("discord.roles"), ",") {
			for _, role := range member.Roles {
				if role == roleID {
					found = true
					count++
					break
				}
			}
			if found {
				break
			}
		}
	}
	memberCount.Set(count)
}

// EventCreate is called whenever an event is created
// It increments eventCount
func EventCreate() {
	eventCount.Dec()
	_, err := globalDB.Exec("INSERT INTO stats VALUES('eventCount', 1) ON DUPLICATE KEY UPDATE value = value + 1;")
	if err != nil {
		log.WithError(err).Error("Failed to update messageCount")
		return
	}
}

// EventRevoke is called whenever an event is revoked
// Decrements eventCount
func EventRevoke() {
	eventCount.Dec()
	_, err := globalDB.Exec("INSERT INTO stats VALUES('eventCount', 0) ON DUPLICATE KEY UPDATE value = value - 1;")
	if err != nil {
		log.WithError(err).Error("Failed to update messageCount")
		return
	}
}

// MessageCreate is called whenever a message is sent
// Increments messageCount for the given server and channel
func MessageCreate(server string, channel string) {
	messageCount.WithLabelValues(server, channel).Inc()
	_, err := globalDB.Exec("INSERT INTO messageCount VALUES(" + server + ", " + channel + ", 1) ON DUPLICATE KEY UPDATE value = value + 1;")
	if err != nil {
		log.WithError(err).Error("Failed to update messageCount")
		return
	}
}

// MessageDelete is called whenever a message is deleted
// Decrements messageCount for the given server and channel
func MessageDelete(server string, channel string) {
	messageCount.WithLabelValues(server, channel).Dec()
	_, err := globalDB.Exec("INSERT INTO messageCount VALUES(" + server + ", " + channel + ", 0) ON DUPLICATE KEY UPDATE value = value - 1;")
	if err != nil {
		log.WithError(err).Error("Failed to update messageCount")
		return
	}
}

func createTables() {
	_, err := globalDB.Exec("CREATE TABLE IF NOT EXISTS stats(name VARCHAR(20) PRIMARY KEY, value INT);")
	if err != nil {
		log.WithError(err).Error("Failed to create table stats")
		return
	}
	_, err = globalDB.Exec("CREATE TABLE IF NOT EXISTS messageCount(server VARCHAR(20), channel VARCHAR(20), value INT, PRIMARY KEY (server, channel));")
	if err != nil {
		log.WithError(err).Error("Failed to create table messageCount")
		return
	}
	_, err = globalDB.Exec("CREATE TABLE IF NOT EXISTS joined(id VARCHAR(20));")
	if err != nil {
		log.WithError(err).Error("Failed to create table joined")
		return
	}
}

func setup(s *discordgo.Session) {
	globalSession = s
	createTables()
	servers := viper.Get("discord.servers").(*config.Servers)
	publicServer, err := s.Guild(servers.PublicServer)
	if err != nil {
		log.WithError(err).Error("Failed to get Public Server guild")
		return
	}
	newMembers := []string{}
	for _, member := range publicServer.Members {
		found := false
		for _, roleID := range strings.Split(viper.GetString("discord.roles"), ",") {
			for _, role := range member.Roles {
				if role == roleID {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			result, err := globalDB.Query("SELECT id FROM joined WHERE id = " + member.User.ID)
			defer result.Close()
			if err != nil {
				log.WithError(err).Error("Failed to get joined")
				return
			}
			if !result.Next() {
				newMembers = append(newMembers, member.User.ID)
				membersJoined.Inc()
			}
			memberCount.Inc()
		}
	}
	if len(newMembers) > 0 {
		_, err = globalDB.Exec("INSERT INTO joined VALUES (" + strings.Join(newMembers, "), (") + ")")
		if err != nil {
			log.WithError(err).Error("Failed to add id to joined")
			return
		}
	}

	result, err := globalDB.Query("SELECT COUNT(*) FROM joined")
	defer result.Close()
	if err != nil {
		log.WithError(err).Error("Failed to get joined")
		return
	}
	if result.Next() {
		var value float64
		if err := result.Scan(&value); err != nil {
			log.WithError(err).Error("Failed to read row")
			return
		}
		membersJoined.Set(value)
	}

	result, err = globalDB.Query("SELECT value FROM stats WHERE name = 'eventCount'")
	defer result.Close()
	if err != nil {
		log.WithError(err).Error("Failed to get stats")
		return
	}
	if result.Next() {
		var value float64
		if err := result.Scan(&value); err != nil {
			log.WithError(err).Error("Failed to read row")
			return
		}
		eventCount.Set(value)
	}

	result, err = globalDB.Query("SELECT server, channel, value FROM messageCount")
	defer result.Close()
	if err != nil {
		log.WithError(err).Error("Failed to get message count")
		return
	}
	for result.Next() {
		var server string
		var channel string
		var value float64
		if err := result.Scan(&server, &channel, &value); err != nil {
			log.WithError(err).Error("Failed to read row")
			return
		}
		messageCount.WithLabelValues(server, channel).Set(value)
	}

}

// CreateExporter should be called when bot is starting
// to set up database tables and start the prometheus exporter http server
func CreateExporter(s *discordgo.Session) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", viper.GetString("mysql.username"), viper.GetString("mysql.password"), viper.GetString("mysql.url"), viper.GetString("prom.dbname")))
	if err != nil {
		log.WithError(err).Error("Failed to connect to db")
		return
	}
	defer func() {
		if db != nil {
			db.Close()
		}
	}()
	globalDB = db
	setup(s)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
