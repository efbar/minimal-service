package tracer

import (
	"context"
	"net"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"

	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/efbar/minimal-service/logging"
)

// TraceObject ...
type TraceObject struct {
}

// Opentracer ...
func (tObj *TraceObject) Opentracer(jaegerURL string, service string, code int, message string, tags []label.KeyValue, l *logging.Logger, envs map[string]string) error {

	if len(jaegerURL) == 0 {
		jaegerURL = "http://localhost:14268/api/traces"
	}
	url, err := url.ParseRequestURI(jaegerURL)
	if err != nil {
		return err
	}

	_, err = net.DialTimeout("tcp", net.JoinHostPort(url.Hostname(), url.Port()), time.Second)
	if err != nil {
		return err
	}

	l.Debug(envs["DEBUG"], "JAEGER_URL", jaegerURL)

	ctx := context.Background()

	flush, err := tObj.initTracer(jaegerURL, service, tags, l, envs)
	defer flush()

	tr := otel.Tracer(service)
	ctx, span := tr.Start(ctx, "request-span")
	if code == 200 {
		span.SetStatus(2, message)
	} else {
		span.SetStatus(1, message)
	}
	defer span.End()

	return err
}

// initTracer creates a new trace provider instance and registers it as global trace provider.
func (tObj *TraceObject) initTracer(jaegerURL string, service string, tags []label.KeyValue, l *logging.Logger, envs map[string]string) (func(), error) {

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
		l.Error(err.Error())
	}
	return flush, err
}
