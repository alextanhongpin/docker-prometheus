package main

import (
	"fmt"
	"net/http"
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
		[]string{"path"},
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
	prometheus.Register(totalRequests)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
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
		totalRequests.WithLabelValues(path).Inc()
		timer.ObserveDuration()
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}
