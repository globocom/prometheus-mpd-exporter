package watcher

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/zencoder/go-dash/v3/mpd"
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

// Periods are ephemeral and need to be collected dynamically to easy clean in memory
type PeriodMetrics struct {
	Registry *prometheus.Registry

	Start   *prometheus.GaugeVec
	BaseURL *prometheus.GaugeVec

	AdaptationSetMimeType *prometheus.GaugeVec
}

const (
	periodSubsystem        = "period"
	adaptationSetSubsystem = "adaptation_set"
)

func NewPeriodMetrics() *PeriodMetrics {
	p := &PeriodMetrics{}

	p.Start = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: periodSubsystem,
			Name:      "start",
			Help:      "Start time of the MPD period in Unix timestamp",
		},
		[]string{"mpd", "period"},
	)
	p.BaseURL = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: periodSubsystem,
			Name:      "base_url",
			Help:      "Base URL of the MPD period",
		},
		[]string{"mpd", "period", "base_url"},
	)

	// AdaptationSet metrics definitions
	p.AdaptationSetMimeType = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: adaptationSetSubsystem,
			Name:      "mime_type",
			Help:      "MIME type of the MPD adaptation set",
		},
		[]string{"mpd", "period", "adaptation_set_index", "mime_type"},
	)

	p.Registry = prometheus.NewRegistry()
	p.Registry.MustRegister(
		p.Start,
		p.BaseURL,
		p.AdaptationSetMimeType,
	)

	return p
}

func collectPeriodMetrics(pm *PeriodMetrics, period *mpd.Period, mpdAlias string) {
	labels := []string{mpdAlias, period.ID}

	if period.Start == nil {
		pm.Start.WithLabelValues(labels...).Set(0)
	} else {
		startDuration := time.Duration(*period.Start)
		pm.Start.WithLabelValues(labels...).Set(startDuration.Seconds())
	}

	for _, baseURL := range period.BaseURL {
		pm.BaseURL.WithLabelValues(appendLabel(labels, baseURL)...).Set(1) // Assuming base URL is always set
	}

	for adaptationSetIndex, adaptationSet := range period.AdaptationSets {
		collectAdaptationSetMetrics(pm, appendLabel(labels, strconv.Itoa(adaptationSetIndex)), adaptationSet)
	}

}

func collectAdaptationSetMetrics(pm *PeriodMetrics, labels []string, adaptationSet *mpd.AdaptationSet) {

	if adaptationSet.MimeType == nil {
		pm.AdaptationSetMimeType.WithLabelValues(appendLabel(labels, "unknown")...).Set(1)
	} else {
		pm.AdaptationSetMimeType.WithLabelValues(appendLabel(labels, *adaptationSet.MimeType)...).Set(1)
	}

	for _, representation := range adaptationSet.Representations {
		collectRepresentationMetrics(pm, labels, representation)
	}
}

func collectRepresentationMetrics(pm *PeriodMetrics, labels []string, representation *mpd.Representation) {

}

func appendLabel(base []string, label string) []string {
	result := make([]string, len(base)+1)
	copy(result, base)
	result[len(base)] = label
	return result
}
