package com.example.mytest;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.InetSocketAddress;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.URLConnection;
import java.util.Scanner;

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
public class MyTestApplication {

	private final Tracer tracer;
	OpenTelemetry openTelemetry;

	// // Tell OpenTelemetry to inject the context in the HTTP headers
	// TextMapSetter<HttpURLConnection> setter = new TextMapSetter<HttpURLConnection>() {
	// 	@Override
	// 	public void set(HttpURLConnection carrier, String key, String value) {
	// 		// Insert the context as Header
	// 		carrier.setRequestProperty(key, value);
	// 	}
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

	public MyTestApplication() {
		openTelemetry = ExampleConfiguration.initOpenTelemetry();
		tracer = openTelemetry.getTracer("io.opentelemetry.example.MyTestApplication", "1.0.1");
	}

	public static void main(String[] args) throws IOException {
		SpringApplication.run(MyTestApplication.class, args);

		MyTestApplication example = new MyTestApplication();

		HttpServer server = HttpServer.create(new InetSocketAddress(8000), 0);
		server.createContext("/test", example.new SimpleHandler());
		server.setExecutor(null); // creates a default executor
		server.start();
	}

	@GetMapping("/1")
	public String hello() {
		myWonderfulUseCase();
		return "Hello world from ANDREWW";
	}

	private void myWonderfulUseCase() {
		// Generate a span
		Span span = this.tracer.spanBuilder("Start my wonderful use case").startSpan();
		span.addEvent("Event 0");
		// execute my use case - here we simulate a wait
		doWork();
		span.addEvent("Event 1");
		doWork();
		span.end();
	}

	private void doWork() {
		try {
			Thread.sleep(1000);
		} catch (InterruptedException e) {
			// do the right thing here
		}
	}

	@GetMapping("/2")
	public String hello2() {
		parentOne();
		return "Hello2 from ANDREWW";
	}

	void parentOne() {
		Span parentSpan = tracer.spanBuilder("parent").startSpan();
		try {
			childOne(parentSpan);
		} finally {
			parentSpan.end();
		}
	}

	void childOne(Span parentSpan) {
		Span childSpan = tracer.spanBuilder("child")
				.setParent(Context.current().with(parentSpan))
				.startSpan();
		try {
			// do stuff
		} finally {
			childSpan.end();
		}
	}

	@GetMapping("/3")
	public String hello3() {
		parentTwo();
		return "Hello 3 from ANDREWW";
	}

	void parentTwo() {
		Span parentSpan = tracer.spanBuilder("parent 3").startSpan();
		parentSpan.setAttribute("http.method", "GET");
		parentSpan.setAttribute("http.url", "..............//////////..........");
		try (Scope scope = parentSpan.makeCurrent()) {
			childTwo();
		} finally {
			parentSpan.end();
		}
	}

	void childTwo() {
		Span childSpan = tracer.spanBuilder("child 3")
				// NOTE: setParent(...) is not required;
				// `Span.current()` is automatically added as the parent
				.startSpan();
		try (Scope scope = childSpan.makeCurrent()) {
			// do stuff
		} finally {
			childSpan.setStatus(StatusCode.ERROR, "Something bad happened!");
			childSpan.end();
		}
	}

	@GetMapping("/4")
	public void hello4() {
		// Tell OpenTelemetry to inject the context in the HTTP headers
		TextMapSetter<HttpURLConnection> setter = new TextMapSetter<HttpURLConnection>() {
			@Override
			public void set(HttpURLConnection carrier, String key, String value) {
				// Insert the context as Header
				carrier.setRequestProperty(key, value);
			}
		};

		try {
			URL url = new URL("http://java2_als:8001/test");
			Span outGoing = tracer.spanBuilder("java2_als/test").setSpanKind(SpanKind.CLIENT).startSpan();
			try (Scope scope = outGoing.makeCurrent()) {
				// Use the Semantic Conventions.
				// (Note that to set these, Span does not *need* to be the current instance in
				// Context or Scope.)
				outGoing.setAttribute(SemanticAttributes.HTTP_METHOD, "GET");
				outGoing.setAttribute(SemanticAttributes.HTTP_URL, url.toString());

				try {
					HttpURLConnection transportLayer = (HttpURLConnection) url.openConnection();
					// Inject the request with the *current* Context, which contains our current
					// Span.
					openTelemetry.getPropagators().getTextMapPropagator().inject(Context.current(), transportLayer,
							setter);
					// Make outgoing call
					transportLayer.setRequestMethod("GET");
					// transportLayer.setRequestProperty("User-Agent", USER_AGENT);
					int responseCode = transportLayer.getResponseCode();
					System.out.println("GET Response Code :: " + responseCode);
					if (responseCode == HttpURLConnection.HTTP_OK) { // success
						BufferedReader in = new BufferedReader(new InputStreamReader(
								transportLayer.getInputStream()));
						String inputLine;
						StringBuffer response = new StringBuffer();

						while ((inputLine = in.readLine()) != null) {
							response.append(inputLine);
						}
						in.close();

						// print result
						System.out.println(response.toString());
					} else {
						System.out.println("GET request not worked");
					}

				} catch (IOException ie) {
					ie.printStackTrace();
				}

			} finally {
				outGoing.end();
			}
		} catch (

		MalformedURLException e) {
			e.printStackTrace();
		}
	}

	public class SimpleHandler implements HttpHandler {
		@Override
		public void handle(HttpExchange httpExchange) throws IOException {
			// Extract the SpanContext and other elements from the request.
			Context extractedContext = openTelemetry.getPropagators().getTextMapPropagator()
					.extract(Context.current(), httpExchange, getter);
			try (Scope scope = extractedContext.makeCurrent()) {
				// Automatically use the extracted SpanContext as parent.
				Span serverSpan = tracer.spanBuilder("GET /4")
						.setSpanKind(SpanKind.SERVER)
						.startSpan();
				try {
					// Add the attributes defined in the Semantic Conventions
					serverSpan.setAttribute(SemanticAttributes.HTTP_METHOD, "GET");
					serverSpan.setAttribute(SemanticAttributes.HTTP_SCHEME, "http");
					serverSpan.setAttribute(SemanticAttributes.HTTP_HOST, "localhost:8080");
					serverSpan.setAttribute(SemanticAttributes.HTTP_TARGET, "/4");
					// Serve the request
					// ...
				} finally {
					serverSpan.end();
				}
			}

			httpExchange.getResponseHeaders().set("Content-type", "text/plain");
			String response = "pong\n";
			httpExchange.sendResponseHeaders(200, response.getBytes().length);
			httpExchange.getResponseBody().write(response.getBytes());
			httpExchange.close();
		}
	}

}
