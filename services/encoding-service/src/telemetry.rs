use lapin::types::{AMQPValue, FieldTable, LongString, ShortString};
use opentelemetry::propagation::{Extractor, Injector};
use opentelemetry::{global, trace::TracerProvider, KeyValue};
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::propagation::TraceContextPropagator;
use opentelemetry_sdk::Resource;
use tracing::warn;

use crate::config::TelemetryConfig;

/// Initializes the OpenTelemetry TracerProvider with an OTLP HTTP exporter.
/// Returns `Some(provider)` on success, `None` if disabled or on failure.
pub fn init_tracer(
    cfg: &TelemetryConfig,
) -> Option<opentelemetry_sdk::trace::TracerProvider> {
    if !cfg.enabled {
        return None;
    }

    let exporter = match opentelemetry_otlp::SpanExporter::builder()
        .with_http()
        .with_endpoint(&cfg.endpoint)
        .build()
    {
        Ok(e) => e,
        Err(e) => {
            warn!(error = %e, "failed to create OTLP exporter, tracing disabled");
            return None;
        }
    };

    let resource = Resource::new(vec![
        KeyValue::new("service.name", cfg.service_name.clone()),
        KeyValue::new("service.version", "1.0.0"),
    ]);

    let provider = opentelemetry_sdk::trace::TracerProvider::builder()
        .with_batch_exporter(exporter, opentelemetry_sdk::runtime::Tokio)
        .with_resource(resource)
        .build();

    global::set_tracer_provider(provider.clone());
    global::set_text_map_propagator(TraceContextPropagator::new());

    Some(provider)
}

/// Creates a `tracing_opentelemetry::OpenTelemetryLayer` from the provider.
pub fn create_otel_layer<S>(
    provider: &opentelemetry_sdk::trace::TracerProvider,
) -> tracing_opentelemetry::OpenTelemetryLayer<S, opentelemetry_sdk::trace::Tracer>
where
    S: tracing::Subscriber + for<'span> tracing_subscriber::registry::LookupSpan<'span>,
{
    let tracer = provider.tracer("encoding-service");
    tracing_opentelemetry::layer().with_tracer(tracer)
}

// ── AMQP Trace Context Propagation ──────────────────────────────────

/// Wraps a mutable `FieldTable` reference for trace context injection.
pub struct AmqpHeaderInjector<'a>(pub &'a mut FieldTable);

impl Injector for AmqpHeaderInjector<'_> {
    fn set(&mut self, key: &str, value: String) {
        self.0.insert(
            ShortString::from(key),
            AMQPValue::LongString(LongString::from(value.as_bytes())),
        );
    }
}

/// Wraps an immutable `FieldTable` reference for trace context extraction.
pub struct AmqpHeaderExtractor<'a>(pub &'a FieldTable);

impl Extractor for AmqpHeaderExtractor<'_> {
    fn get(&self, key: &str) -> Option<&str> {
        self.0.inner().get(&ShortString::from(key)).and_then(|v| {
            if let AMQPValue::LongString(s) = v {
                std::str::from_utf8(s.as_bytes()).ok()
            } else {
                None
            }
        })
    }

    fn keys(&self) -> Vec<&str> {
        self.0.inner().keys().map(|k| k.as_str()).collect()
    }
}

/// Injects the current span's trace context into a new `FieldTable`.
pub fn inject_context() -> FieldTable {
    use tracing_opentelemetry::OpenTelemetrySpanExt;

    let mut headers = FieldTable::default();
    let cx = tracing::Span::current().context();
    global::get_text_map_propagator(|propagator| {
        propagator.inject_context(&cx, &mut AmqpHeaderInjector(&mut headers));
    });
    headers
}

/// Extracts trace context from AMQP delivery headers into an OpenTelemetry `Context`.
pub fn extract_context(headers: &FieldTable) -> opentelemetry::Context {
    global::get_text_map_propagator(|propagator| {
        propagator.extract(&AmqpHeaderExtractor(headers))
    })
}
