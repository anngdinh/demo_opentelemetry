package main

import (
	// "encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	// resty "github.com/go-resty/resty/v2"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	// "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/exporters/jaeger"
)

type Result struct {
	Message string
}

const (
	service     = "trace-factorial"
	environment = "production"
	id          = 2
)

func TracerProvider() (*tracesdk.TracerProvider, error) {
	url := "http://allinone:14268/api/traces"
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func main() {
	_, err := TracerProvider()
	if err != nil {
		log.Fatal(err)
	}
	router := gin.Default()
	router.Use(otelgin.Middleware("microservice-2"))
	{
		router.GET("/pong", func(c *gin.Context) {
			ctx := c.Request.Context()
			span := trace.SpanFromContext(otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(c.Request.Header)))
			defer span.End()

			c.IndentedJSON(200, gin.H{
				"message": "pong",
			})
		})
	}
	router.Run(":8088")

}
