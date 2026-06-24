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

	// Temporal Go SDK emits latency histograms in seconds (unit="s").
	// Boundaries here must also be in seconds.
	// Kind filter is required so this view does not override gauges/counters.
	secondsView := metric.NewView(
		metric.Instrument{
			Name:  "*",
			Kind:  metric.InstrumentKindHistogram,
			Scope: instrumentation.Scope{Name: "temporal-sdk-go"},
		},
		metric.Stream{
			Aggregation: metric.AggregationExplicitBucketHistogram{
				Boundaries: []float64{
					0.001, 0.005, 0.010, 0.025, 0.050,
					0.100, 0.250, 0.500,
					1, 2, 5, 10, 30, 60,
				},
			},
		},
	)

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(1*time.Second))),
		metric.WithView(secondsView),
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
