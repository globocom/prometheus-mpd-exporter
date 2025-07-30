package watcher

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/zencoder/go-dash/v3/mpd"
)

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

func collectMPDMetrics(mpegMPD *mpd.MPD, mpdAlias string) error {
	if mpegMPD.AvailabilityStartTime == nil {
		mpdAvailabilityStartTime.WithLabelValues(mpdAlias).Set(0)
	} else {
		parsedAvailabilityStartTime, err := time.Parse(time.RFC3339, *mpegMPD.AvailabilityStartTime)
		if err != nil {
			return err
		}

		mpdAvailabilityStartTime.WithLabelValues(mpdAlias).Set(float64(parsedAvailabilityStartTime.Unix()))
	}

	if mpegMPD.PublishTime == nil {
		mpdPublishTime.WithLabelValues(mpdAlias).Set(0)
	} else {
		parsedPublishTime, err := time.Parse(time.RFC3339, *mpegMPD.PublishTime)
		if err != nil {
			return err
		}

		mpdPublishTime.WithLabelValues(mpdAlias).Set(float64(parsedPublishTime.Unix()))
	}

	mpdPeriods.WithLabelValues(mpdAlias).Set(float64(len(mpegMPD.Periods)))

	if len(mpegMPD.Periods) > 0 {
		lastPeriodID := mpegMPD.Periods[len(mpegMPD.Periods)-1].ID
		lastPeriodIDInt, _ := strconv.Atoi(lastPeriodID)
		mpdLastPeriod.WithLabelValues(mpdAlias).Set(float64(lastPeriodIDInt))
	} else {
		mpdLastPeriod.WithLabelValues(mpdAlias).Set(0)
	}

	pm := NewPeriodMetrics()

	for _, period := range mpegMPD.Periods {
		collectPeriodMetrics(pm, period, mpdAlias)
	}

	LastPeriodMetrics[mpdAlias].Store(pm)

	return nil
}
