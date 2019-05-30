package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

func main() {
	exporter, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "demo",
	})

	if err != nil {
		log.Fatal(err)
	}

	view.RegisterExporter(exporter)

	// Measures description.
	var (
		// A string by which the measure will be referred to by.
		name        = "my.org/measures/latency"
		description = "The latency in milliseconds"
		unit        = "ms"
	)
	// https://opencensus.io/stats/measure/
	mLatencyMs := stats.Float64(name, description, unit)
	mLines := stats.Int64("lines_in", "The number of lines processed", "1")
	mBytesIn := stats.Int64("bytes_in", "The number of bytes received", "By")

	// Views.
	keyMethod, err := tag.NewKey("keymethod")
	if err != nil {
		log.Fatal(err)
	}
	latencyView := &view.View{
		Name:        "myapp/latency",
		Measure:     mLatencyMs,
		Description: "The distribution of the latencies",
		TagKeys:     []tag.Key{keyMethod},
		Aggregation: view.Distribution(0, 25, 100, 200, 400, 800, 10000),
	}

	lineCountView := &view.View{
		Name:        "myapp/lines",
		Measure:     mLatencyMs,
		Description: "The number of lines that were received",
		TagKeys:     []tag.Key{keyMethod},
		Aggregation: view.Count(),
	}
	view.Register(latencyView, lineCountView)

	go func() {
		for {
			ctx := context.Background()
			stats.Record(ctx,
				mLatencyMs.M(float64(rand.Intn(20))),
				mLines.M(int64(rand.Intn(100))+100),
				mBytesIn.M(rand.Int63n(10000)))
			<-time.After(time.Millisecond * time.Duration(1+rand.Intn(400)))
		}
	}()

	http.Handle("/metrics", exporter)
	log.Println("listening to port *:2112")
	log.Fatal(http.ListenAndServe(":2112", nil))
}
