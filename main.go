package main

import (
	"context"
	"fmt"
	"log"
	"mp/fib/pkg/fib"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName    = "mp/fib"
	instrumentationVersion = "0.1.0"

	otlpEndpointUrl = "http://localhost:4318/v1/traces" // http not https
)

func Resource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(fib.ServiceName),
		semconv.ServiceVersion(fib.ServiceVersion),
	)
}

func InstallExportPipeline() (func(context.Context) error, error) {
	// stdoutExporter := check(stdouttrace.New(stdouttrace.WithPrettyPrint()))

	otlpExporter := check(otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpointURL(otlpEndpointUrl),
	))

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(otlpExporter),
		// sdktrace.WithBatcher(stdoutExporter),
		sdktrace.WithResource(Resource()),
	)
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
}

func main() {
	ctx := context.Background()

	// Registers a tracer Provider globally.
	shutdown := check(InstallExportPipeline())

	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	tracer := otel.GetTracerProvider().Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(instrumentationVersion),
		trace.WithSchemaURL(semconv.SchemaURL),
	)

	ctx, span := tracer.Start(ctx, "main")
	defer span.End()

	var n uint

	// n = <-fib.Channel(ctx, 10)
	// fmt.Println(n)

	n = fib.Recurse(ctx, 10) //, fib.WithSimpleMemoization())
	fmt.Println(n)

	span.SetAttributes(attribute.Int("n", int(n)))
}

func check[T any](thing T, err error) T {
	if err != nil {
		panic(err)
	}
	return thing
}
