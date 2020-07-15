package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	member_count = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "member_count",
		Help: "The total number of members in the server",
	})
	members_joined = promauto.NewCounter(prometheus.CounterOpts{
		Name: "members_joined",
		Help: "The total number of members to have ever joined the server",
	})
	event_count = promauto.NewCounter(prometheus.CounterOpts{
		Name: "event_count",
		Help: "The total number of events ran",
	})
	message_count = promauto.NewCounter(prometheus.CounterOpts{
		Name: "message_count",
		Help: "The total number of messages sent in the server",
	})
)

func MemberJoin() {
	member_count.Inc()
	members_joined.Inc()
}

func MemberLeave() {
	member_count.Dec()
}

func EventCreate() {
	event_count.Inc()
}

func MessageCreate() {
	message_count.Inc()
}

func CreateExporter() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
