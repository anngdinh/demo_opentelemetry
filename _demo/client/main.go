package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	resty "github.com/go-resty/resty/v2"
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
	service     = "trace-demo"
	environment = "production"
	id          = 1
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
	resty := resty.New()

	router := gin.Default()
	router.Use(otelgin.Middleware("microservice-1"))
	{

		router.GET("/ping", func(c *gin.Context) {
			result := Result{}
			req := resty.R().SetHeader("Content-Type", "application/json")
			ctx := req.Context()
			span := trace.SpanFromContext(ctx)

			defer span.End()

			otel.GetTextMapPropagator().Inject(c.Request.Context(), propagation.HeaderCarrier(req.Header))
			resp, err := req.Get("http://java_als:8000/test")

			if err != nil {
				fmt.Println("---------error get /pong-----------")
				// log.Fatal(err)
			}

			json.Unmarshal([]byte(resp.String()), &result)
			fmt.Println(resp)
			fmt.Println(result)
			c.IndentedJSON(200, gin.H{
				"message": result.Message,
				// "message2": resp,
			})
		})

		router.GET("/ping2", func(c *gin.Context) {
			result := Result{}
			req := resty.R().SetHeader("Content-Type", "application/json")
			ctx := req.Context()
			span := trace.SpanFromContext(ctx)

			defer span.End()

			otel.GetTextMapPropagator().Inject(c.Request.Context(), propagation.HeaderCarrier(req.Header))
			resp, err := req.Get("http://factorial_als:8088/5")

			if err != nil {
				fmt.Println("---------error get /5-----------")
				// log.Fatal(err)
			}

			json.Unmarshal([]byte(resp.String()), &result)
			c.IndentedJSON(200, gin.H{
				"message": result.Message,
			})
		})
	}
	router.Run(":8080")

}