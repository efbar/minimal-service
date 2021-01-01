package tracer

import (
	"context"
	"fmt"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"

	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TraceObject ...
type TraceObject struct {
}

// Opentracer ...
func (tObj *TraceObject) Opentracer(url string, service string, tags []label.KeyValue) error {
	ctx := context.Background()

	flush, err := tObj.initTracer(url, service, tags)
	defer flush()

	tr := otel.Tracer(service)
	ctx, span := tr.Start(ctx, "span")
	defer span.End()

	return err
}

// initTracer creates a new trace provider instance and registers it as global trace provider.
func (tObj *TraceObject) initTracer(jaegerURL string, service string, tags []label.KeyValue) (func(), error) {

	if len(jaegerURL) == 0 {
		jaegerURL = "http://localhost:14268/api/traces"
	}
	_, err := url.ParseRequestURI(jaegerURL)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("JAEGER_URL", jaegerURL)

	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(jaegerURL),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: service,
			Tags:        tags,
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		fmt.Println(err)
	}
	return flush, err
}
