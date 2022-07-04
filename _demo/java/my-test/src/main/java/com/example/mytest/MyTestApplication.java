package com.example.mytest;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.StatusCode;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;

@SpringBootApplication
@RestController
public class MyTestApplication {

	private final Tracer tracer;
	OpenTelemetry openTelemetry;

	public MyTestApplication() {
		openTelemetry = ExampleConfiguration.initOpenTelemetry();
		tracer = openTelemetry.getTracer("io.opentelemetry.example.MyTestApplication", "1.0.1");
	}

	public static void main(String[] args) {
		SpringApplication.run(MyTestApplication.class, args);
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
}
