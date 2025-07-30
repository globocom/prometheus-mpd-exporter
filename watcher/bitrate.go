package watcher

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/zencoder/go-dash/v3/mpd"
	"k8s.io/utils/ptr"
)

// bitrate metrics

var (
	bitrateMetrics = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: bitrateSubsystem,
			Name:      "segments_size_bytes_total",
			Help:      "Counter for the total size of all segments in bytes",
		},
		[]string{"mpd", "representation_id", "mime_type"},
	)
)

type lastSegmentKey struct {
	MpdAlias         string
	RepresentationID string
	MimeType         string
}

var lastSegmentRecords = &sync.Map{}

func collectBitrateMetrics(mpegMPD *mpd.MPD, mpdAlias, url string) error {
	parts := strings.Split(url, "/")
	baseURL := strings.Join(parts[:len(parts)-1], "/") + "/"

	for _, period := range mpegMPD.Periods {
		collectBitratePeriodMetrics(period, mpdAlias, baseURL)
	}
	return nil
}

func collectBitratePeriodMetrics(period *mpd.Period, mpdAlias, baseURL string) {

	if len(period.BaseURL) > 0 {
		baseURL = period.BaseURL[0]
	}

	for _, adaptationSet := range period.AdaptationSets {
		collectBitrateAdaptationSetMetrics(adaptationSet, mpdAlias, baseURL)
	}
}

func collectBitrateAdaptationSetMetrics(adaptationSet *mpd.AdaptationSet, mpdAlias, baseURL string) {
	for _, representation := range adaptationSet.Representations {
		collectBitrateRepresentationMetrics(adaptationSet, representation, mpdAlias, baseURL)
	}
}

func collectBitrateRepresentationMetrics(adaptationSet *mpd.AdaptationSet, representation *mpd.Representation, mpdAlias, baseURL string) {
	if adaptationSet.SegmentTemplate == nil || adaptationSet.SegmentTemplate.SegmentTimeline == nil {
		return
	}
	if adaptationSet.MimeType == nil {
		return
	}
	if representation.ID == nil {
		return
	}

	mediaPath := strings.ReplaceAll(*adaptationSet.SegmentTemplate.Media, "$RepresentationID$", *representation.ID)

	lastUsageKey := lastSegmentKey{
		MpdAlias:         mpdAlias,
		RepresentationID: *representation.ID,
		MimeType:         *adaptationSet.MimeType,
	}

	var lastSegmentTime uint64
	lastSegmentValue, found := lastSegmentRecords.Load(lastUsageKey)
	if found {
		lastSegmentTime = lastSegmentValue.(uint64)
	}

	segments := unwindSegments(adaptationSet.SegmentTemplate.SegmentTimeline.Segments)
	for _, segment := range segments {
		if segment.StartTime == nil || *segment.StartTime <= lastSegmentTime {
			continue
		}

		segmentPath := strings.ReplaceAll(mediaPath, "$Time$", fmt.Sprintf("%d", *segment.StartTime))
		collectSegmentSize(mpdAlias, *representation.ID, *adaptationSet.MimeType, baseURL+segmentPath)
		lastSegmentRecords.Store(lastUsageKey, *segment.StartTime)
	}
}

func collectSegmentSize(mpdAlias, representationID, mimeType, url string) {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Printf("Error fetching segment size for %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	bitrateMetrics.WithLabelValues(mpdAlias, representationID, mimeType).Add(float64(resp.ContentLength))
}

func unwindSegments(segments []*mpd.SegmentTimelineSegment) []*mpd.SegmentTimelineSegment {
	unwindSegments := []*mpd.SegmentTimelineSegment{}

	for _, segmentsTemplate := range segments {
		startTime := int64(*segmentsTemplate.StartTime)
		duration := int64(segmentsTemplate.Duration)

		if segmentsTemplate.RepeatCount == nil {
			segmentsTemplate.RepeatCount = ptr.To(int(0))
		}

		for i := 0; i < int(*segmentsTemplate.RepeatCount+1); i++ {

			unwindSegments = append(unwindSegments, &mpd.SegmentTimelineSegment{
				StartTime: ptr.To(uint64(startTime)),
				Duration:  uint64(duration),
			})

			startTime = startTime + duration
		}
	}

	return unwindSegments
}
