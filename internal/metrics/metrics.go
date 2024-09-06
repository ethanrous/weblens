package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "weblens_http_requests_total",
			Help: "The total number of http requests",
		},
	)
	RequestsTimer = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Time it takes to process requests",
			Buckets: []float64{0.0001, 0.001, 0.01, 0.1, 1, 10},
		},
		[]string{"handler", "method"},
	)
	MediaProcessTime = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "weblens_media_process_time_seconds",
			Help: "The time it takes to process a media",
		},
	)
	BusyWorkerGuage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "weblens_busy_worker_count_gauge",
			Help: "Number of workers curently executing a task",
		},
	)
)
