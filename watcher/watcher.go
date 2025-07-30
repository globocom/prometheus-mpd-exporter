package watcher

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/zencoder/go-dash/v3/mpd"
)

var LastPeriodMetrics = map[string]*atomic.Value{}

func Init(mpdAlias, url string) {
	LastPeriodMetrics[mpdAlias] = &atomic.Value{}

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

	pm := NewPeriodMetrics()

	for _, period := range mpegMPD.Periods {
		collectPeriodMetrics(pm, period, mpdAlias)
	}

	LastPeriodMetrics[mpdAlias].Store(pm)

	return nil
}
