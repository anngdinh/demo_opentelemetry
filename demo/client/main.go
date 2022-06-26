package main

import (
	"fmt"
	"sync"
	"time"

	"context"
	"math/rand"
	"net/http"
	// "net/url"
	// "os"
	"github.com/gin-gonic/gin"
	"strconv"

	"github.com/opentracing/opentracing-go/ext"
	// "github.com/opentracing/opentracing-go/log"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	xhttp "github.com/yurishkuro/opentracing-tutorial/go/lib/http"
	logrus "github.com/sirupsen/logrus"
)


func main() {
	// initDB()

	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			// HTTPHeaders:
			LocalAgentHostPort: "allinone:6831",
			// CollectorEndpoint: "allinone:6831",
		},
	}

	router := gin.New()
	router.Use(CORSMiddleware())

	router.GET("/factorial/:num", func(c *gin.Context) {

		tracer, closer, _ := cfg.New(
			"demo_seminar",
			config.Logger(jaeger.StdLogger),
		)
		opentracing.SetGlobalTracer(tracer)
		defer closer.Close()

		span := tracer.StartSpan("demo-factorial")
		span.SetTag("function", "main")
		defer span.Finish()

		ctx := opentracing.ContextWithSpan(context.Background(), span)

		num, _ := strconv.Atoi(c.Param("num"))
		span.SetTag("value", num)

		fac, err := factorial(ctx, num)

		if err != nil {
			logrus.Error("Error in client/main function !")
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			fmt.Println("factorial = ", fac)
			c.String(http.StatusOK, "fac (%d) = %d ", num, fac)
		}

	})

	router.GET("/fibonacci/:num", func(c *gin.Context) {

		tracer, closer, _ := cfg.New(
			"demo_seminar",
			config.Logger(jaeger.StdLogger),
		)
		opentracing.SetGlobalTracer(tracer)
		defer closer.Close()

		span := tracer.StartSpan("demo-fibonacci")
		span.SetTag("function", "main")
		defer span.Finish()

		span.SetBaggageItem("count_fib", c.Param("0"))

		ctx := opentracing.ContextWithSpan(context.Background(), span)

		num, _ := strconv.Atoi(c.Param("num"))
		span.SetTag("value", num)

		fac := fibonacci(ctx, num)

		fmt.Println("fibonacci = ", fac)

		c.String(http.StatusOK, "fibonacci(%d) = %d ", num, fac)
	})

	router.GET("/serial/:num", func(c *gin.Context) {
		tracer, closer, _ := cfg.New(
			"demo_seminar",
			config.Logger(jaeger.StdLogger),
		)
		opentracing.SetGlobalTracer(tracer)
		defer closer.Close()

		span := tracer.StartSpan("demo-serial")
		span.SetTag("function", "main")
		defer span.Finish()

		ctx := opentracing.ContextWithSpan(context.Background(), span)

		num, _ := strconv.Atoi(c.Param("num"))
		span.SetTag("value", num)

		fib := fibonacci(ctx, num)
		fac, _ := factorial(ctx, num)

		fmt.Println("fibonacci = ", fib)
		fmt.Println("factorial = ", fac)

		c.String(http.StatusOK, "factorial(%d) = %d\nfibonacci(%d) = %d", num, fac, num, fib)
	})

	router.GET("/concurrency/:num", func(c *gin.Context) {
		tracer, closer, _ := cfg.New(
			"demo_seminar",
			config.Logger(jaeger.StdLogger),
		)
		opentracing.SetGlobalTracer(tracer)
		defer closer.Close()

		span := tracer.StartSpan("demo-concurrency")
		span.SetTag("function", "main")
		defer span.Finish()

		ctx := opentracing.ContextWithSpan(context.Background(), span)

		num, _ := strconv.Atoi(c.Param("num"))
		span.SetTag("value", num)

		ch := make(chan int)
		go fac_concurrency(ctx, num, ch)
		go fib_concurrency(ctx, num, ch)
		fib, fac := <-ch, <-ch // receive from ch

		fmt.Println("fibonacci = ", fib)
		fmt.Println("factorial = ", fac)

		c.String(http.StatusOK, "factorial(%d) = %d\nfibonacci(%d) = %d", num, fac, num, fib)
	})

	

	

	router.Run(":8080")
}
