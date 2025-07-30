package watcher

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
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

	bitrateMetrics := viper.GetBool("bitrate-metrics")

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

	if bitrateMetrics {
		err := collectBitrateMetrics(mpegMPD, mpdAlias, url)
		if err != nil {
			return err
		}
	}

	return nil
}
