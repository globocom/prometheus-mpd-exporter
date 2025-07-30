package watcher

import (
	"log"
	"time"
)

func Init(mpdAlias, url string) {
	go func() {
		for {
			watcherIter(mpdAlias, url)
			time.Sleep(time.Second)
		}
	}()
}

func watcherIter(mpdAlias, url string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in watcher for %s: %v", mpdAlias, r)
		}
	}()

	log.Printf("Starting watcher for MPD host: %s at %s", mpdAlias, url)
}
