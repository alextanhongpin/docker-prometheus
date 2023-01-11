package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// responseWriter allows us to capture the http status code through the middleware.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var (
	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of get requests",
		},
		[]string{"method", "path"},
	)

	responseStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "response_status",
			Help: "Status of HTTP response",
		},
		[]string{"status"},
	)

	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_response_time_seconds",
			Help: "Duration of HTTP requests",
		},
		[]string{"path"},
	)
)

func init() {
	prometheus.MustRegister(totalRequests)
	prometheus.MustRegister(responseStatus)
	prometheus.MustRegister(httpDuration)
	prometheus.MustRegister(NewCollector("myapp"))
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", prometheusMiddleware(http.HandlerFunc(indexHandler)))
	mux.Handle("/greet", prometheusMiddleware(http.HandlerFunc(greetHandler)))
	mux.Handle("/metrics", promhttp.Handler())
	fmt.Println("listening to port *:2112. press ctrl + c to cancel")
	http.ListenAndServe(":2112", mux)
}

// NOTE: We don't need to implement this manually as it already exists as part of the promhttp.Handler()
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		responseStatus.WithLabelValues(strconv.Itoa(rw.statusCode)).Inc()
		totalRequests.WithLabelValues(r.Method, path).Inc()
		timer.ObserveDuration()
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

// NewCollector returns a collector that exports metrics about current version
// information.
// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Gauge:~:text=NewGaugeFunc%20is%20a%20good%20way%20to%20create%20an%20%E2%80%9Cinfo%E2%80%9D%20style%20metric%20with%20a%20constant%20value%20of%201.%20Example%3A%20https%3A//github.com/prometheus/common/blob/8558a5b7db3c84fa38b4766966059a7bd5bfa2ee/version/info.go%23L36%2DL56
//
// This will appear as myapp_build_info
//
// Output:
// myapp_build_info{branch="master", goversion="go1.19.4", instance="host.docker.internal:2112", job="myapp", revision="e69834f", version="0.0.1"}
func NewCollector(program string) prometheus.Collector {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: program,
			Name:      "build_info",
			Help: fmt.Sprintf(
				"A metric with a constant '1' value labeled by version, revision, branch, and goversion from which %s was built.",
				program,
			),
			ConstLabels: prometheus.Labels{
				"version":   os.Getenv("VERSION"),
				"revision":  os.Getenv("REVISION"),
				"branch":    os.Getenv("BRANCH"),
				"goversion": runtime.Version(),
			},
		},
		func() float64 { return 1 },
	)
}
