package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

func Setup() (func(), error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}
	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
	)
	otel.SetMeterProvider(provider)

	shutdown := func() {
		_ = provider.Shutdown(context.Background())
	}
	return shutdown, nil
}
