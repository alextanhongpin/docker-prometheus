## Using Prometheus

```bash
$ go get github.com/prometheus/client_golang/prometheus
$ go get github.com/prometheus/client_golang/prometheus/promauto
$ go get github.com/prometheus/client_golang/prometheus/promhttp
```

## Using OpenCensus

The package `go.opencensus.io/exporter/stats/prometheus` has been replaced with the `contrib.go.opencensus.io/exporter/prometheus`.
```bash
$ go get -u go.opencensus.io/...
$ go get -u contrib.go.opencensus.io/exporter/prometheus
```

Opentracing

```
$ go get -u -v contrib.go.opencensus.io/exporter/zipkin
$ go get -u -v github.com/openzipkin/zipkin-go
```
