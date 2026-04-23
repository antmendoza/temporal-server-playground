package loadgen

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
)

const collectorEndpoint = "http://localhost:4318"

// NewMetricsHandler pushes SDK metrics to the OTEL collector via OTLP HTTP.
// serviceName is used to distinguish worker vs starter in Prometheus labels.
func NewMetricsHandler(serviceName string) client.MetricsHandler {
	ctx := context.Background()

	exp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(collectorEndpoint),
	)
	if err != nil {
		log.Fatalln("failed to create OTLP exporter:", err)
	}

	// Explicit histogram buckets for Temporal latency metrics (milliseconds)
	view := metric.NewView(
		metric.Instrument{
			Name:  "*",
			Scope: instrumentation.Scope{Name: "temporal-sdk-go"},
		},
		metric.Stream{
			Aggregation: metric.AggregationExplicitBucketHistogram{
				Boundaries: []float64{0, 50, 100, 200, 500, 1000, 2000, 5000, 10000},
			},
		},
	)

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(5*time.Second))),
		metric.WithView(view),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	log.Printf("SDK metrics pushing to %s (service=%s)", collectorEndpoint, serviceName)
	return opentelemetry.NewMetricsHandler(opentelemetry.MetricsHandlerOptions{
		Meter: mp.Meter("temporal-sdk-go"),
	})
}
