package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		reg       = prometheus.NewRegistry()
		mediumRes = promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:                "kubeconeu2022_demo",
			Help:                "Values observed during the story.",
			SparseBucketsFactor: 2,
			ConstLabels:         map[string]string{"type": "med_res"},
		})
		lowRes = promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:                "kubeconeu2022_demo",
			Help:                "Values observed during the story.",
			SparseBucketsFactor: 4,
			ConstLabels:         map[string]string{"type": "low_res"},
		})
		mediumResLatenciesStage1 = []float64{
			0.6, 0.7, // (0.5, 1]
			2.1, 2.5, 3.5, // (2, 4]
			20, 25, 30, // (16, 32]
		}
		mediumResLatenciesStage2 = []float64{
			1.5,  // (1, 2]
			6, 7, // (4, 8]
			20, 25, 30, // (16, 32]
		}
		mediumResLatencies = []float64{
			0.6, 0.7, // (0.5, 1]
			1.5,           // (1, 2]
			2.1, 2.5, 3.5, // (2, 4]
			6, 7, // (4, 8]
			20, 25, 30, // (16, 32]
		}
		lowResLatencies = []float64{
			1.5, 2, 3, 3.5, // (1, 4]
			5, 6, 7, // (4, 16]
			33.33, // (16, 64]
		}
	)

	observe := func(h prometheus.Histogram, latencies []float64) {
		for _, latency := range latencies {
			h.Observe(latency)
		}
	}
	resetMediumRes := func() {
		reg.Unregister(mediumRes)
		mediumRes = promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:                "kubeconeu2022_demo",
			Help:                "Values observed during the story.",
			SparseBucketsFactor: 2,
			ConstLabels:         map[string]string{"type": "med_res"},
		})
	}

	go func() {
		observe(mediumRes, mediumResLatenciesStage1)
		<-time.After(2 * time.Minute)
		resetMediumRes()
		observe(mediumRes, mediumResLatenciesStage2)
		<-time.After(2 * time.Minute)
		resetMediumRes()
		observe(mediumRes, mediumResLatencies)
		observe(lowRes, lowResLatencies)
		<-time.After(2 * time.Minute)

		for {
			<-time.After(20 * time.Millisecond)
			for _, latency := range mediumResLatencies {
				mediumRes.Observe(latency)
			}
			for _, latency := range lowResLatencies {
				lowRes.Observe(latency)
			}
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Println("Serving metrics, SIGTERM to abort…")
	http.ListenAndServe(":8080", nil)
}
