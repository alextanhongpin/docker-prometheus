package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"contrib.go.opencensus.io/exporter/zipkin"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

func main() {
	// Configure exporter to export traces to Zipkin.
	localEndpoint, err := openzipkin.NewEndpoint("go-quickstart", "127.0.0.1:5454")
	if err != nil {
		log.Fatal(err)
	}
	reporter := zipkinHTTP.NewReporter("http://localhost:9411/api/v2/spans")
	ze := zipkin.NewExporter(reporter, localEndpoint)
	trace.RegisterExporter(ze)

	// Configure 100% sample rate, otherwise, few traces will be sampled.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// Create a span with the background context, making this the parent sapn. A span must be closed.
	ctx, span := trace.StartSpan(context.Background(), "main")
	defer span.End()

	for i := 0; i < 10; i++ {
		doWork(ctx)
	}
}

func doWork(ctx context.Context) {
	_, span := trace.StartSpan(ctx, "doWork")
	defer span.End()

	fmt.Println("busy doing work")
	time.Sleep(80 * time.Millisecond)
	buf := bytes.NewBuffer([]byte{0xFF, 0x00, 0x00, 0x00})
	num, err := binary.ReadVarint(buf)
	if err != nil {
		// Set status upon error.
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: err.Error(),
		})
	}
	// Annotate our span to capture metadata abour our operation.
	span.Annotate([]trace.Attribute{
		trace.Int64Attribute("bytes to int", num),
	}, "Invoking doWork")
	time.Sleep(20 * time.Millisecond)
}
