package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"nmapshot/config"
	_ "nmapshot/docs" // This line is crucial for swagger to find the docs
	"nmapshot/handler"
	"nmapshot/middleware"
)

//go:generate go tool swag init -g main.go

// initTracer initializes an OTLP exporter, and configures the corresponding trace provider.
func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceName("nmapshot"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider, nil
}

// @title Nmap Shot API
// @version 1.0
// @description REST API service to execute Nmap scans and return XML results.

// @license.name MIT
// @license.url https://github.com/gpy/nmapshot/blob/main/LICENSE

// @host localhost:8082
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, relying on system environment variables")
	}

	// Initialize OpenTelemetry
	tp, err := initTracer()
	if err != nil {
		log.Printf("Failed to initialize OpenTelemetry: %v", err)
	} else {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}()
	}

	// Load allowed ports configuration
	allowedPorts, err := config.LoadAllowedPorts()
	if err != nil {
		log.Fatalf("Failed to load allowed ports config: %v", err)
	}
	log.Printf("Allowed ports: %s", allowedPorts)

	r := gin.Default()
	r.Use(otelgin.Middleware("nmapshot"))

	// Swagger configuration
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/swagger/index.html")
	})
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/swagger/index.html")
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API Routes (protected by API key)
	scanHandler := handler.NewScanHandler(allowedPorts)
	v1 := r.Group("/api/v1")
	v1.Use(middleware.APIKeyAuth())
	{
		v1.POST("/scan", scanHandler.Handle)
	}

	port := config.LoadRestAPIPort()
	log.Printf("Server starting on :%s...\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
