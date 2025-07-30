package watcher

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/zencoder/go-dash/v3/mpd"
)

func Init(mpdAlias, url string) {
	mpdInfo.WithLabelValues(mpdAlias, url).Set(1) // Set the MPD info gauge to 1

	go func() {
		for {
			err := watcherIter(mpdAlias, url)
			if err != nil {
				log.Printf("Error in watcher for %s: %v", mpdAlias, err)
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func watcherIter(mpdAlias, url string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in watcher for %s: %v", mpdAlias, r)
		}
	}()

	resp, err := http.Get(url) // Simulate a request to the MPD host
	if err != nil {
		return errors.Wrap(err, "failed to get MPD host")
	}
	defer resp.Body.Close()

	mpdFetchStatus.WithLabelValues(mpdAlias, strconv.Itoa(resp.StatusCode)).Inc()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected status code %d from MPD host %s", resp.StatusCode, mpdAlias)
	}

	mpegMPD, err := mpd.Read(resp.Body)
	if err != nil {
		return err
	}

	if err := collectMPDMetrics(mpegMPD, mpdAlias); err != nil {
		return err
	}

	return nil
}

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

	for _, period := range mpegMPD.Periods {
		collectPeriodMetrics(period, mpdAlias)
	}

	return nil
}

func collectPeriodMetrics(period *mpd.Period, mpdAlias string) {
	if period.Start == nil {
		mpdPeriodStart.WithLabelValues(mpdAlias, period.ID).Set(0)
	} else {
		startDuration := time.Duration(*period.Start)
		mpdPeriodStart.WithLabelValues(mpdAlias, period.ID).Set(startDuration.Seconds())
	}

	for _, baseURL := range period.BaseURL {
		mpdPeriodBaseURL.WithLabelValues(mpdAlias, period.ID, baseURL).Set(1) // Assuming base URL is always set
	}

	for _, adaptationSet := range period.AdaptationSets {
		collectAdaptationSetMetrics(adaptationSet, mpdAlias, period.ID)
	}
}

func collectAdaptationSetMetrics(adaptationSet *mpd.AdaptationSet, mpdAlias, periodID string) {
	if adaptationSet.MimeType == nil {
		mpdAdaptationSetMimeType.WithLabelValues(mpdAlias, periodID, "unknown").Set(1)
	} else {
		mpdAdaptationSetMimeType.WithLabelValues(mpdAlias, periodID, *adaptationSet.MimeType).Set(1)
	}

}
