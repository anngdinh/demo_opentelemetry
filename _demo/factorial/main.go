package main

import (
	// "encoding/json"
	"context"
	"log"
	// "net/http"
	"strconv"

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

	// resty "github.com/go-resty/resty/v2"
)

var tracer trace.Tracer

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
		router.GET("/:num", func(c *gin.Context) {
			num := c.Param("num")
			savedCtx := c.Request.Context()
			defer func() {
				c.Request = c.Request.WithContext(savedCtx)
			}()

			ctx := otel.GetTextMapPropagator().Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))
			ctx, span := tracer.Start(ctx, "spanname=" + num)
			defer span.End()

			span.SetAttributes(attribute.String("func", "fact_main"))
			// span.SetAttributes(attribute.String("value", num))

			span.AddEvent("start calc ...")
			// num_atoi, _ := strconv.Atoi(num)
			// ctx = context.TODO()
			// fac_small(ctx, num_atoi)
			span.AddEvent("finish calc ...")
			
			// c.Request = c.Request.WithContext(ctx)
			// c.Next()
			// c.IndentedJSON(200, gin.H{
			// 	"message": fac,
			// })
		})
	}
	router.Run(":8088")

}

func fac_small(ctx context.Context, n int) (int, error) {
	ctx, childSpan := tracer.Start(ctx, "span_name="+strconv.Itoa(n))
	// childSpan.SetAttributes(attribute.String("func", "fact_small"))
	// childSpan.SetAttributes(attribute.String("value", n))
	defer childSpan.End()

	if n <= 1 {
		// if init := span.BaggageItem("init"); init == "10" {
		// 	log.Error("error from factorial/fac_small with num =  " + strconv.Itoa(n))
		// 	span.SetAttributes(attribute.String("error", true))
		// 	return 0, errors.New("Can't solve in fac_small")
		// }
		return 1, nil
	}
	// ctx2 := tracer.Start(context.Background(), span)

	// client := http.Client{}
	// req, err := http.NewRequest("GET", "http://java2_als:8001/test" + strconv.Itoa(n - 1), nil)
	// if err != nil {
	// 	//Handle Error
	// }

	// req.Header = http.Header{
	// 	"Content-Type": {"application/json"},
	// }
	// otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// _, err2 := client.Do(req)
	// if err2 != nil {
	// 	//Handle Error
	// }
	
	// resty := resty.New()
	// req := resty.R().SetHeader("Content-Type", "application/json")
	// ctx = req.Context()
	// span := trace.SpanFromContext(ctx)

	// defer span.End()

	// otel.GetTextMapPropagator().Inject(c.Request.Context(), propagation.HeaderCarrier(req.Header))
	// resp, err := req.Get("http://factorial_als:8088/5")

	// if err != nil {
	// 	fmt.Println("---------error get /5-----------")
	// 	// log.Fatal(err)
	// }

	// childSpan.SetAttributes(attribute.String("result", res))
	return -99, nil
}
