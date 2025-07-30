package watcher

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const metricsNamespace = "mpd"

// MPD metrics definitions
var (
	mpdInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "info",
			Help:      "Information about the MPD host",
		},
		[]string{"mpd", "url"},
	)

	mpdFetchStatus = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "fetch_status_code_total",
			Help:      "Counter for the status of MPD fetch operations",
		},
		[]string{"mpd", "status"},
	)

	mpdPeriods = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "periods",
			Help:      "Gauge for the number of periods in the MPD",
		},
		[]string{"mpd"},
	)

	mpdLastPeriod = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "last_period",
			Help:      "Number of the last period in the MPD",
		},
		[]string{"mpd"},
	)

	mpdAvailabilityStartTime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "availability_start_time",
			Help:      "Availability start time of the MPD in Unix timestamp",
		},
		[]string{"mpd"},
	)

	mpdPublishTime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "publish_time",
			Help:      "Publish time of the MPD in Unix timestamp",
		},
		[]string{"mpd"},
	)
)

const periodSubsystem = "period"

// Period metrics definitions
var (
	mpdPeriodStart = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: periodSubsystem,
			Name:      "start",
			Help:      "Start time of the MPD period in Unix timestamp",
		},
		[]string{"mpd", "period"},
	)

	mpdPeriodBaseURL = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: periodSubsystem,
			Name:      "base_url",
			Help:      "Base URL of the MPD period",
		},
		[]string{"mpd", "period", "base_url"},
	)
)

// AdaptationSet metrics definitions

const adaptationSetSubsystem = "adaptation_set"

var (
	mpdAdaptationSetMimeType = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: adaptationSetSubsystem,
			Name:      "mime_type",
			Help:      "MIME type of the MPD adaptation set",
		},
		[]string{"mpd", "period", "mime_type"},
	)
)
