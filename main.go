package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/globocom/prometheus-mpd-exporter/watcher"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initFlags()

	mpdHosts := viper.GetStringMapString("mpd_hosts")

	if len(mpdHosts) == 0 {
		log.Fatal("No MPD hosts provided. Please set the 'mpd-hosts' flag with at least one MPD host.")
	}

	for alias, url := range mpdHosts {
		log.Printf("Initializing watcher for MPD host: %s at %s", alias, url)
		// Initialize the watcher for each MPD host
		watcher.Init(alias, url)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Prometheus MPD Exporter\n"))
		w.Write([]byte("Visit /metrics for metrics\n"))
		w.Write([]byte("Visit https://github.com/globocom/prometheus-mpd-exporter to learn more\n"))
	})

	mux.Handle("/metrics", promhttp.Handler())

	port := viper.GetString("port")
	log.Printf("Starting Prometheus MPD Exporter on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))

}

func initFlags() {
	flag.Int("port", 8888, "Port to listen on")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()
}
