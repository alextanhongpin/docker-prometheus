package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func recordMetrics() chan struct{} {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				opsProcessed.Inc()
				time.Sleep(1 * time.Second)
			}
		}
	}()
	return done
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func main() {
	done := recordMetrics()
	defer close(done)
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("listening to port *:2112. press ctrl + c to cancel")
	http.ListenAndServe(":2112", nil)
}
