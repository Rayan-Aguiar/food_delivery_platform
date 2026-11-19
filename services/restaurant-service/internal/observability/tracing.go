package observability

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/event"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type TracingConfig struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// SetupTracing segue o mesmo padrao do auth-service. O endpoint do exporter vem
// das variaveis padrao do OTel (OTEL_EXPORTER_OTLP_ENDPOINT etc.).
func SetupTracing(ctx context.Context, cfg TracingConfig) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if !cfg.Enabled {
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

// NewHTTPHandler cria o span pai de cada request. O span nasce com o nome
// operation e, ao final, o otelhttp o renomeia pelo formatter quando r.Pattern
// foi preenchido pelo mux; por isso nao pode haver middleware que clone a
// request entre este handler e o mux. Sem rota (ex.: 404), fica operation.
func NewHTTPHandler(handler http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(
		withHTTPRoute(handler),
		operation,
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			if r.Pattern != "" {
				return r.Pattern
			}
			return operation
		}),
	)
}

// withHTTPRoute adiciona http.route ao span: o otelhttp so le r.Pattern no
// inicio da request, quando o mux ainda nao resolveu a rota.
func withHTTPRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if r.Pattern != "" {
			trace.SpanFromContext(r.Context()).SetAttributes(semconv.HTTPRoute(r.Pattern))
		}
	})
}

func NewMongoMonitor() *event.CommandMonitor {
	return otelmongo.NewMonitor()
}

// AnnotateMenuValidation adiciona ao span da request o resultado da validacao,
// permitindo filtrar traces de pedidos rejeitados sem abrir o payload.
func AnnotateMenuValidation(ctx context.Context, restaurantID string, itemCount int, allValid bool) {
	trace.SpanFromContext(ctx).SetAttributes(
		attribute.String("restaurant.id", restaurantID),
		attribute.Int("menu.validation.item_count", itemCount),
		attribute.Bool("menu.validation.all_valid", allValid),
	)
}
