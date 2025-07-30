package watcher

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zencoder/go-dash/v3/mpd"
)

// Periods are ephemeral and need to be collected dynamically to easy clean in memory
type PeriodMetrics struct {
	Registry *prometheus.Registry

	Start   *prometheus.GaugeVec
	BaseURL *prometheus.GaugeVec

	AdaptationSetMimeType *prometheus.GaugeVec

	RepresentationHeight    *prometheus.GaugeVec
	RepresentationWidth     *prometheus.GaugeVec
	RepresentationBandwidth *prometheus.GaugeVec
	RepresentationCodecs    *prometheus.GaugeVec
}

const (
	metricsNamespace        = "mpd"
	periodSubsystem         = "period"
	adaptationSetSubsystem  = "adaptation_set"
	representationSubsystem = "representation"
	bitrateSubsystem        = "bitrate"
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
		[]string{"mpd", "period", "adaptation_set", "mime_type"},
	)

	// Representation metrics definitions
	p.RepresentationHeight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: representationSubsystem,
			Name:      "height",
			Help:      "Height of the MPD representation in pixels",
		},
		[]string{"mpd", "period", "adaptation_set", "representation"},
	)

	p.RepresentationWidth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: representationSubsystem,
			Name:      "width",
			Help:      "Width of the MPD representation in pixels",
		},
		[]string{"mpd", "period", "adaptation_set", "representation"},
	)

	p.RepresentationBandwidth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: representationSubsystem,
			Name:      "bandwidth",
			Help:      "Bandwidth of the MPD representation in bits per second",
		},
		[]string{"mpd", "period", "adaptation_set", "representation"},
	)

	p.RepresentationCodecs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: representationSubsystem,
			Name:      "codecs",
			Help:      "Codecs of the MPD representation",
		},
		[]string{"mpd", "period", "adaptation_set", "representation", "codecs"},
	)

	p.Registry = prometheus.NewRegistry()
	p.Registry.MustRegister(
		p.Start,
		p.BaseURL,
		p.AdaptationSetMimeType,
		p.RepresentationHeight,
		p.RepresentationWidth,
		p.RepresentationBandwidth,
		p.RepresentationCodecs,
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

		adaptationSetID := strconv.Itoa(adaptationSetIndex)
		if adaptationSet.ID != nil {
			adaptationSetID = *adaptationSet.ID
		}

		collectAdaptationSetMetrics(pm, appendLabel(labels, adaptationSetID), adaptationSet)
	}

}

func collectAdaptationSetMetrics(pm *PeriodMetrics, labels []string, adaptationSet *mpd.AdaptationSet) {

	if adaptationSet.MimeType == nil {
		pm.AdaptationSetMimeType.WithLabelValues(appendLabel(labels, "unknown")...).Set(1)
	} else {
		pm.AdaptationSetMimeType.WithLabelValues(appendLabel(labels, *adaptationSet.MimeType)...).Set(1)
	}

	for index, representation := range adaptationSet.Representations {

		representationID := strconv.Itoa(index)
		if representation.ID != nil {
			representationID = *representation.ID
		}

		collectRepresentationMetrics(pm, appendLabel(labels, representationID), representation)
	}

}

func collectRepresentationMetrics(pm *PeriodMetrics, labels []string, representation *mpd.Representation) {
	if representation.Height != nil {
		pm.RepresentationHeight.WithLabelValues(labels...).Set(float64(*representation.Height))
	}

	if representation.Width != nil {
		pm.RepresentationWidth.WithLabelValues(labels...).Set(float64(*representation.Width))
	}

	if representation.Bandwidth != nil {
		pm.RepresentationBandwidth.WithLabelValues(labels...).Set(float64(*representation.Bandwidth))
	}

	if representation.Codecs != nil {
		pm.RepresentationCodecs.WithLabelValues(appendLabel(labels, *representation.Codecs)...).Set(1)
	}

}

func appendLabel(base []string, label string) []string {
	result := make([]string, len(base)+1)
	copy(result, base)
	result[len(base)] = label
	return result
}
