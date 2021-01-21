package prometheus

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"database/sql"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/utils"
	"github.com/bwmarrin/discordgo"

	// Needed for mysql
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

var (
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

// MemberJoinLeave should be called every time a member joins or leaves.
func MemberJoinLeave() {
	servers := viper.Get("discord.servers").(*config.Servers)
	publicServer, err := utils.GetGuildPreview(globalSession, servers.PublicServer)
	if err != nil {
		log.WithError(err).Error("Failed to get Public Server guild")
		return
	}
	membersJoined.Set(float64(publicServer.ApproximateMemberCount))
}

// EventCreate is called whenever an event is created
// It increments eventCount
func EventCreate() {
	eventCount.Dec()
	_, err := globalDB.Exec("INSERT INTO stats VALUES('eventCount', 1) ON CONFLICT (server, channel) DO UPDATE SET value = excluded.value + 1;")
	if err != nil {
		log.WithError(err).Error("Failed to update messageCount")
		return
	}
}

// EventRevoke is called whenever an event is revoked
// Decrements eventCount
func EventRevoke() {
	eventCount.Dec()
	_, err := globalDB.Exec("INSERT INTO stats VALUES('eventCount', 0) ON CONFLICT (server, channel) DO UPDATE SET value = excluded.value - 1;")
	if err != nil {
		log.WithError(err).Error("Failed to update messageCount")
		return
	}
}

// MessageCreate is called whenever a message is sent
// Increments messageCount for the given server and channel
func MessageCreate(server string, channel string) {
	messageCount.WithLabelValues(server, channel).Inc()
	_, err := globalDB.Exec("INSERT INTO messageCount VALUES(" + server + ", " + channel + ", 1) ON CONFLICT (server, channel) DO UPDATE SET value = excluded.value + 1;")
	if err != nil {
		log.WithError(err).Error("Failed to update messageCount")
		return
	}
}

// MessageDelete is called whenever a message is deleted
// Decrements messageCount for the given server and channel
func MessageDelete(server string, channel string) {
	messageCount.WithLabelValues(server, channel).Dec()
	_, err := globalDB.Exec("INSERT INTO messageCount VALUES(" + server + ", " + channel + ", 0) ON CONFLICT (server, channel) DO UPDATE SET value = excluded.value - 1;")
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
}

func setup(s *discordgo.Session) {
	globalSession = s
	createTables()
	MemberJoinLeave()

	result, err := globalDB.Query("SELECT value FROM stats WHERE name = 'eventCount'")
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
	db, err := sql.Open("postgres",
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			viper.GetString("sql.host"),
			viper.GetInt("sql.port"),
			viper.GetString("sql.username"),
			viper.GetString("sql.password"),
			viper.GetString("prom.dbname"),
		),
	)
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
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
