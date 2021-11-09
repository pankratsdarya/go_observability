package main

import (
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	Namespace = "randmetrics"

	LabelMethod = "method"
	LabelStatus = "status"
)

type app struct {
	latencyHistogram,
	firstLengthHistogram *prometheus.HistogramVec

	lineCounter prometheus.Counter

	secondLengthGauge prometheus.Gauge
}

func (a *app) processHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	firstLen := rand.Intn(100)
	secondLen := rand.Intn(500)

	line := r.URL.Query().Get("line")

	a.firstLengthHistogram.With(prometheus.Labels{LabelStatus: "OK"}).Observe(float64(firstLen))

	a.secondLengthGauge.Set(float64(secondLen))

	defer func() {
		a.latencyHistogram.With(prometheus.Labels{LabelMethod: r.Method}).Observe(sinceInMilliseconds(startTime))

		a.lineCounter.Inc()
	}()

	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // имитация работы

	writeResponse(w, http.StatusOK, strings.ToUpper(line))
}

func (a *app) Init() error {
	// prometheus type: histogram
	a.latencyHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: Namespace,
		Name:      "latency",
		Help:      "The distribution of the latencies",
		Buckets:   []float64{0, 25, 50, 75, 100, 200, 400, 600, 800, 1000, 2000, 4000, 6000},
	}, []string{LabelMethod})

	// prometheus type: histogram
	a.firstLengthHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: Namespace,
		Name:      "first_rand_length",
		Help:      "First random lenght",
		// длины: [>=0, >=10, >=20, >=30, >=40, >=50, >=60, >=70, >=80, >=90, >=100]
		Buckets: []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
	}, []string{LabelStatus})

	// prometheus type: counter
	a.lineCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "lines_in",
		Help:      "The number of lines from standard input",
	})

	// prometheus type: gauge
	a.secondLengthGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Name:      "second_rand_length",
		Help:      "Second random length",
	})

	prometheus.MustRegister(a.latencyHistogram)
	prometheus.MustRegister(a.firstLengthHistogram)
	prometheus.MustRegister(a.lineCounter)
	prometheus.MustRegister(a.secondLengthGauge)

	return nil
}

func (a *app) Serve() error {
	mux := http.NewServeMux()
	mux.Handle("/process", http.HandlerFunc(a.processHandler)) // /process?line=текст+тут
	mux.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe("0.0.0.0:9000", mux)
}

func main() {
	a := app{}

	if err := a.Init(); err != nil {
		log.Fatal(err)
	}

	if err := a.Serve(); err != nil {
		log.Fatal(err)
	}
}

func sinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}

func writeResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(message))
	_, _ = w.Write([]byte("\n"))
}
