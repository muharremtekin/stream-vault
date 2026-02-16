use std::fmt;

use chrono::Utc;
use opentelemetry::trace::TraceContextExt;
use serde_json::{Map, Value};
use tracing::{Event, Subscriber};
use tracing_opentelemetry::OpenTelemetrySpanExt;
use tracing_subscriber::fmt::format::{self, FormatEvent, FormatFields};
use tracing_subscriber::fmt::FmtContext;
use tracing_subscriber::registry::LookupSpan;

/// JSON log formatter that injects `service`, `trace_id`, and `span_id` from
/// the active OpenTelemetry span context into every log line.
pub struct ServiceJsonFormat {
    service_name: String,
}

impl ServiceJsonFormat {
    pub fn new(service_name: impl Into<String>) -> Self {
        Self {
            service_name: service_name.into(),
        }
    }
}

struct JsonVisitor(Map<String, Value>);

impl JsonVisitor {
    fn new() -> Self {
        Self(Map::new())
    }
}

impl tracing::field::Visit for JsonVisitor {
    fn record_debug(&mut self, field: &tracing::field::Field, value: &dyn fmt::Debug) {
        self.0
            .insert(field.name().into(), Value::String(format!("{:?}", value)));
    }

    fn record_str(&mut self, field: &tracing::field::Field, value: &str) {
        self.0
            .insert(field.name().into(), Value::String(value.to_owned()));
    }

    fn record_i64(&mut self, field: &tracing::field::Field, value: i64) {
        self.0
            .insert(field.name().into(), Value::Number(value.into()));
    }

    fn record_u64(&mut self, field: &tracing::field::Field, value: u64) {
        self.0
            .insert(field.name().into(), Value::Number(value.into()));
    }

    fn record_f64(&mut self, field: &tracing::field::Field, value: f64) {
        if let Some(n) = serde_json::Number::from_f64(value) {
            self.0.insert(field.name().into(), Value::Number(n));
        }
    }

    fn record_bool(&mut self, field: &tracing::field::Field, value: bool) {
        self.0.insert(field.name().into(), Value::Bool(value));
    }
}

impl<S, N> FormatEvent<S, N> for ServiceJsonFormat
where
    S: Subscriber + for<'a> LookupSpan<'a>,
    N: for<'a> FormatFields<'a> + 'static,
{
    fn format_event(
        &self,
        _ctx: &FmtContext<'_, S, N>,
        mut writer: format::Writer<'_>,
        event: &Event<'_>,
    ) -> fmt::Result {
        let meta = event.metadata();

        let mut visitor = JsonVisitor::new();
        event.record(&mut visitor);

        let mut obj = Map::new();
        obj.insert(
            "timestamp".into(),
            Value::String(Utc::now().to_rfc3339_opts(chrono::SecondsFormat::Millis, true)),
        );
        obj.insert(
            "level".into(),
            Value::String(meta.level().as_str().to_lowercase()),
        );
        obj.insert(
            "target".into(),
            Value::String(meta.target().to_owned()),
        );
        obj.insert(
            "service".into(),
            Value::String(self.service_name.clone()),
        );

        // Extract OTel trace context from the current span.
        let current_span = tracing::Span::current();
        let otel_ctx = current_span.context();
        let otel_span = otel_ctx.span();
        let sc = otel_span.span_context();
        if sc.is_valid() {
            obj.insert(
                "trace_id".into(),
                Value::String(sc.trace_id().to_string()),
            );
            obj.insert(
                "span_id".into(),
                Value::String(sc.span_id().to_string()),
            );
        }

        // Extract message field.
        if let Some(msg) = visitor.0.remove("message") {
            obj.insert("message".into(), msg);
        }

        // Remaining event fields.
        for (k, v) in visitor.0 {
            obj.insert(k, v);
        }

        let json_str = serde_json::to_string(&obj).map_err(|_| fmt::Error)?;
        writeln!(writer, "{}", json_str)
    }
}
