package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"log/slog" // Add this for slog.Logger

	deepobservergo "github.com/dquang0504/deep-observer/libs/deep-observer-go"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace" // Add this import for trace.Tracer
)

// initTelemetry initializes OpenTelemetry Tracing & Metrics
// It returns a shutdown function to be deferred in main
func initTelemetry(ctx context.Context) (func(context.Context) error, error) {
	// Connect to OTEL Collector (running at Docker localhost:4318)
	// Resource identifies who sends the data
	res, _ := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("sample-service"),
		),
	)

	// Configure Trace Exporter (sending trace via HTTP)
	traceExporter, _ := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())

	// TracerProvider manages the state of tracing
	// WithBatcher: Optimizes network by sending traces in batches instead of one by one
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp) // Register globally

	// Configure Metric Exporter (sending metrics via HTTP)
	metricExporter, _ := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure())

	// MeterProvider manages metrics
	// PeriodicReader: Pushes metrics every interval (default 60s)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp) // Register globally

	// Configure Log Exporter (sending logs via HTTP)
	logExporter, _ := otlploghttp.New(ctx, otlploghttp.WithInsecure())

	// LoggerProvider manages logs
	// WWithProcessor: BatchProcessor behaves like trace batcher
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)), sdklog.WithResource(res))
	global.SetLoggerProvider(lp) //Register globally

	// Cleanup function to shut down connections cleanly
	shutdown := func(ctx context.Context) error {
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
		_ = lp.Shutdown(ctx)
		return nil
	}
	return shutdown, nil
}

func simulateWork(ctx context.Context, tracer trace.Tracer, logger *slog.Logger) {
	// 1. Parent Span: Process Order
	ctx, span := tracer.Start(ctx, "ProcessOrder")
	defer span.End()

	orderID := uuid.New().String()
	logger.InfoContext(ctx, "Processing new order", "order_id", orderID)

	// Simulate work duration
	time.Sleep(time.Millisecond * 100)

	// 2. Child Span: Check Inventory
	func() {
		ctx, span := tracer.Start(ctx, "CheckInventory")
		defer span.End()
		logger.InfoContext(ctx, "Checking inventory database", "order_id", orderID)
		time.Sleep(time.Millisecond * 50)
		// 10% chance failure
		if time.Now().UnixNano()%10 == 0 {
			err := fmt.Errorf("inventory connection timeout")
			span.RecordError(err)
			span.SetStatus(codes.Error, "Inventory check failed")
			logger.ErrorContext(ctx, "Inventory check failed", "error", err, "order_id", orderID)
			return
		}
		span.AddEvent("Inventory reserved")
	}()

	// 3. Child Span: Process Payment
	func() {
		ctx, span := tracer.Start(ctx, "ProcessPayment")
		defer span.End()
		logger.InfoContext(ctx, "Contacting payment gateway", "order_id", orderID)
		time.Sleep(time.Millisecond * 200)
		span.AddEvent("Payment authorized")
		logger.InfoContext(ctx, "Payment successful", "amount", 99.99, "currency", "USD")
	}()

	span.SetStatus(codes.Ok, "Order processed")
	logger.InfoContext(ctx, "Order processing complete", "order_id", orderID, "status", "success")
}

func main() {
	ctx := context.Background()
	//Initialize Telemetry
	shutdown, err := initTelemetry(ctx)
	if err != nil {
		log.Fatalf("failed to init telemetry: %v", err)
	}
	defer shutdown(ctx) //ensure data is sent before app exits

	// Create OTLP Logger (slog compatible)
	logger := otelslog.NewLogger("sample-service")
	// Create Tracer & Meter
	tracer := otel.Tracer("sample-service")
	meter := otel.Meter("sample-service")

	// Create a metric counter
	runCounter, _ := meter.Int64Counter("sample_service.runs_total", metric.WithDescription("Total number of service runs"))
	runCounter.Add(ctx, 1)

	// --- SIMULATION START ---
	// Loop to generate multiple traces for better visualization
	fmt.Println("Simulating traffic... (Press Ctrl+C to stop)")
	for i := 0; i < 5; i++ {
		simulateWork(ctx, tracer, logger)
		fmt.Printf("Finished request %d/5\n", i+1)
		time.Sleep(time.Second * 1)
	}
	// --- SIMULATION END ---

	// Pointing to Control Plane API (default port 8090)
	client := deepobservergo.NewClient("http://localhost:8090")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare a Deployment Event (Just one at the end to show integration)
	event := deepobservergo.Event{
		ID:            uuid.New().String(),
		SchemaVersion: "1.0",
		EventType:     "deployment",
		ServiceName:   "sample-service",
		Environment:   "dev",
		Title:         "Deploying Sample Service v2.0-Simulation",
		OccurredAt:    time.Now(),
		Version:       "2.0.0",
		CommitHash:    "abcd12345",
		Actor:         "william-dang",
	}

	logger.InfoContext(ctx, "Sending deployment event", "service_name", event.ServiceName)
	err = client.SendEvent(ctx, event)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to send event", "error", err)
	} else {
		logger.InfoContext(ctx, "Successfully sent event to Deep-Observer!")
	}
}
