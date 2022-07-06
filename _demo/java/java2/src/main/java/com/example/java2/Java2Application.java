package com.example.java2;

import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.InetSocketAddress;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.SpanKind;
import io.opentelemetry.api.trace.StatusCode;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;
import io.opentelemetry.context.propagation.TextMapGetter;
import io.opentelemetry.context.propagation.TextMapSetter;
import io.opentelemetry.semconv.trace.attributes.SemanticAttributes;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

@SpringBootApplication
@RestController
public class Java2Application {

	private final Tracer tracer;
	OpenTelemetry openTelemetry;

	// // Tell OpenTelemetry to inject the context in the HTTP headers
	// TextMapSetter<HttpURLConnection> setter = new
	// TextMapSetter<HttpURLConnection>() {
	// @Override
	// public void set(HttpURLConnection carrier, String key, String value) {
	// // Insert the context as Header
	// carrier.setRequestProperty(key, value);
	// }
	// };

	TextMapGetter<HttpExchange> getter = new TextMapGetter<>() {
		@Override
		public String get(HttpExchange carrier, String key) {
			if (carrier.getRequestHeaders().containsKey(key)) {
				return carrier.getRequestHeaders().get(key).get(0);
			}
			return null;
		}

		@Override
		public Iterable<String> keys(HttpExchange carrier) {
			return carrier.getRequestHeaders().keySet();
		}
	};

	public Java2Application() {
		openTelemetry = ExampleConfiguration.initOpenTelemetry();
		tracer = openTelemetry.getTracer("io.opentelemetry.example.Java2Application", "1.0.1");
	}

	public static void main(String[] args) throws IOException {
		SpringApplication.run(Java2Application.class, args);

		Java2Application example = new Java2Application();

		HttpServer server = HttpServer.create(new InetSocketAddress(8001), 0);
		server.createContext("/test", example.new SimpleHandler());
		server.setExecutor(null); // creates a default executor
		server.start();
	}

	public class SimpleHandler implements HttpHandler {
		@Override
		public void handle(HttpExchange httpExchange) throws IOException {
			// Extract the SpanContext and other elements from the request.
			Context extractedContext = openTelemetry.getPropagators().getTextMapPropagator()
					.extract(Context.current(), httpExchange, getter);
			try (Scope scope = extractedContext.makeCurrent()) {
				// Automatically use the extracted SpanContext as parent.
				Span serverSpan = tracer.spanBuilder("GET /test")
						.setSpanKind(SpanKind.SERVER)
						.startSpan();
				try {
					// Add the attributes defined in the Semantic Conventions
					serverSpan.setAttribute(SemanticAttributes.HTTP_METHOD, "GET");
					serverSpan.setAttribute(SemanticAttributes.HTTP_SCHEME, "http");
					serverSpan.setAttribute(SemanticAttributes.HTTP_HOST, "localhost:8080");
					serverSpan.setAttribute(SemanticAttributes.HTTP_TARGET, "/resource");
					// Serve the request
					// ...
				} finally {
					serverSpan.end();
				}
			}

			httpExchange.getResponseHeaders().set("Content-type", "text/plain");
			String response = "pong from java2\n";
			httpExchange.sendResponseHeaders(200, response.getBytes().length);
			httpExchange.getResponseBody().write(response.getBytes());
			httpExchange.close();
		}
	}

}
