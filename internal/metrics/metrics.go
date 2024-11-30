package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry         *prometheus.Registry
	connectedPlayers prometheus.Gauge
)

func init() {
	registry = prometheus.NewRegistry()
	connectedPlayers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "scribblers",
		Name:      "connected_players",
		Help:      "The amount of connected players (active websocket connections)",
	})

	registry.MustRegister(connectedPlayers)
}

func TrackPlayerConnect() {
	connectedPlayers.Inc()
}

func TrackPlayerDisconnect() {
	connectedPlayers.Dec()
}

func SetupRoute(registerFunc func(http.HandlerFunc)) {
	registerFunc(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP)
}
