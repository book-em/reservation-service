package util

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type SpanPair struct {
	Ctx  context.Context
	Span trace.Span
}

type Telemetry struct {
	// During tests, the tracer is not set up, so we silently ignore tracing.
	// This has to be done manually (i.e. don't call tracer methods, don't touch
	// spans etc.)
	tracerReady bool
	Tracer      trace.Tracer

	SpanStack []SpanPair
}

var TEL Telemetry

func (t *Telemetry) Init(ctx context.Context, serviceName, deploymentEnvironment string) func(context.Context) error {
	// [1] Init tracer
	{
		exp, err := otlptracehttp.New(ctx)
		if err != nil {
			log.Fatalf("Failed to create an OTLP HTTP exporter: %v", err)
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName(serviceName),
				semconv.DeploymentEnvironment(deploymentEnvironment),
			)),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.TraceContext{})

		t.Tracer = otel.Tracer(serviceName)
		t.tracerReady = true
		return tp.Shutdown
	}
}

func (t *Telemetry) Push(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if t.tracerReady {
		newCtx, span := t.Tracer.Start(ctx, name, trace.WithAttributes(attrs...))
		t.SpanStack = append(t.SpanStack, SpanPair{Ctx: newCtx, Span: span})
	} else {
		newCtx := ctx
		var span trace.Span
		t.SpanStack = append(t.SpanStack, SpanPair{Ctx: newCtx, Span: span})
	}
}

func (t *Telemetry) Pop() {
	top := t.SpanStack[len(t.SpanStack)-1]
	if t.tracerReady {
		top.Span.End()
	}
	t.SpanStack = t.SpanStack[:len(t.SpanStack)-1]
}

func (t *Telemetry) Top() SpanPair {
	return t.SpanStack[len(t.SpanStack)-1]
}

func (t *Telemetry) Ctx() context.Context {
	if len(t.SpanStack) > 0 {
		return t.Top().Ctx
	} else {
		return context.Background() // Ehh...
	}
}

func (t *Telemetry) Event(msg string, err error) {
	// Logging
	{
		if err == nil {
			log.Println(msg)
		} else {
			log.Printf("%s: %v\n", msg, err)
		}
	}

	// Tracing
	if t.tracerReady && len(t.SpanStack) > 0 {
		span := t.SpanStack[len(t.SpanStack)-1].Span

		if err == nil {
			span.AddEvent(msg)
		} else {
			span.AddEvent(msg, trace.WithAttributes(
				attribute.String("error.message", err.Error()),
				attribute.Bool("error", true),
			))
			span.SetStatus(codes.Error, err.Error())
		}
	}
}

func (t *Telemetry) Eventf(msg string, err error, a ...any) {
	msgFinal := fmt.Sprintf(msg, a...)
	t.Event(msgFinal, err)
}

func (t *Telemetry) SetAttrib(kv ...attribute.KeyValue) {
	if t.tracerReady {
		t.Top().Span.SetAttributes(kv...)
	}
}

func (t *Telemetry) SetUser(id uint) {
	if t.tracerReady {
		t.Top().Span.SetAttributes(attribute.String("user.id", fmt.Sprintf("%d", id)))
	}
}

func (t *Telemetry) Inject(outgoingRequest *http.Request) {
	otel.GetTextMapPropagator().Inject(t.Ctx(), propagation.HeaderCarrier(outgoingRequest.Header))
}
