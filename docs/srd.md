# DEEP-OBSERVER SYSTEM REQUIREMENTS DOCUMENT (SRD)
## Version: v7.0 (OTel-Native Pivot)
## Status: Approved / Definitive

---

## 1. Introduction
Deep-Observer is a reusable, production-oriented observability platform designed to provide deep visibility into modern distributed systems. The platform focuses on collecting, processing, storing, and visualizing telemetry signals including logs, metrics, traces, and operational events.

This SRD defines the OTel-native architectural design for Deep-Observer. It eliminates proprietary databases and custom HTTP API endpoints for telemetry or event collection, aligning the platform entirely with open-source industry standards (OpenTelemetry, Loki, Prometheus, Tempo, and Grafana).

---

## 2. Problem Statement & Goals
### 2.1 Problem Statement
Modern microservice-based systems are complex and distributed. Traditional monitoring systems provide surface-level metrics but fail to correlate system metrics/traces with system lifecycle events (such as code deployments, configuration changes, or active incidents). 

Commercial APM solutions solve this but introduce high costs and vendor lock-in. Deep-Observer provides a vendor-neutral, cost-effective correlation platform that connects application telemetry with lifecycle events to reduce MTTR (Mean Time to Resolution).

### 2.2 Primary Goals
*   **Zero Proprietary Code Ingestion:** Standardize all telemetry and operational event collection on the OpenTelemetry Protocol (OTLP).
*   **Event Correlation:** Correlate metric spikes and traces with system lifecycle events (e.g., "how did deployment v1.2 affect latency?").
*   **Minimal Application Friction:** Provide clean, non-intrusive initialization helpers and configurations via a CLI bootstrapping tool.
*   **SRE Best Practices:** Implement Tail-Based Sampling, RED/USE metrics generation, and Log-to-Trace correlation.

---

## 3. High-Level Architecture
Deep-Observer runs as an independent observability pipeline alongside application services.

```mermaid
graph TD
    subgraph "Application Services"
        App[Go / Node.js App] -->|OTLP gRPC/HTTP| Collector
    end

    subgraph "CI/CD & Operational Tools"
        CICD[GitHub Actions / Script] -->|OTLP HTTP Log| Collector
    end

    subgraph "Deep-Observer Pipeline"
        Collector[OTel Collector] -->|Metrics| Prom[Prometheus]
        Collector -->|Traces| Tempo[Tempo]
        Collector -->|Logs & Events| Loki[Loki]
    end

    subgraph "Visualization & Alerting"
        Grafana[Grafana] -->|Query| Prom
        Grafana -->|Query| Tempo
        Grafana -->|Query| Loki
        Alert[Alertmanager] <--|Alerting| Prom
    end
```

### 3.1 Telemetry Flow
1.  **Instrumentation:** Application services utilize official OpenTelemetry SDKs to collect traces, metrics, and logs.
2.  **Export:** Telemetry is exported over standard OTLP (gRPC on port `4317` or HTTP on port `4318`) to the OTel Collector.
3.  **Routing & Processing:** The OTel Collector filters, batches, samples, and exports data to respective backend databases:
    *   **Metrics:** Prometheus (pull-based scraping from the Collector's Prometheus Exporter).
    *   **Logs & Operational Events:** Grafana Loki (via OTLP HTTP receiver).
    *   **Traces:** Grafana Tempo (via OTLP gRPC receiver).
4.  **Visualization:** Grafana queries all backends to provide a unified dashboard view with metric-trace-log correlation.

---

## 4. Telemetry & Event Payload Specifications

### 4.1 Structured Application Logs
All application logs MUST be output as structured JSON lines conforming to OpenTelemetry semantic conventions.

*   **Required Fields:**
    *   `timestamp`: RFC 3339 formatted ISO-8601 string.
    *   `message`: Human-readable log payload.
    *   `severity_text`: Severity level (DEBUG, INFO, WARN, ERROR, FATAL).
    *   `service.name`: Logical name of the service emitting the log.
    *   `environment`: Running environment (e.g., `dev`, `staging`, `prod`).
    *   `trace_id`: 32-character hex string representing the active W3C trace ID.
    *   `span_id`: 16-character hex string representing the active span ID.

Example payload:
```json
{
  "timestamp": "2026-05-20T12:00:00.123Z",
  "message": "database query completed successfully",
  "severity_text": "INFO",
  "service.name": "order-service",
  "environment": "prod",
  "trace_id": "4fd0c9d0b7a64a5a8b0fa77b9d1c0a12",
  "span_id": "a3c1b2d4e5f60789",
  "duration_ms": 12.4
}
```

### 4.2 Operational Events (Deployments, Incidents, Maintenance)
Operational events are represented as OTel Log Records with specialized attributes, sent directly via OTLP. This completely replaces the custom database table schemas.

*   **Attribute Schema:**
    *   `event.name`: The event identifier (e.g., `deployment`, `incident`, `maintenance`).
    *   `event.status`: The state of the event (`started`, `completed`, `failed`, `mitigated`).
    *   `service.name`: The name of the service affected.
    *   `environment`: Target environment (`dev`, `staging`, `prod`).
    *   `event.title`: Brief descriptive title (used for Grafana Annotations).
    *   `event.description`: Detailed description (markdown supported).
    *   `deployment.version`: Application version/tag (for deployment events).
    *   `deployment.commit`: Git commit SHA.
    *   `actor.name`: System/User triggering the event (e.g., `GitHub Actions`, `john-doe`).

Example event payload (sent as an OTLP Log Record):
```json
{
  "timestamp": "2026-05-20T12:05:00Z",
  "message": "deployment completed",
  "severity_text": "INFO",
  "attributes": {
    "event.name": "deployment",
    "event.status": "completed",
    "event.title": "order-service v1.4.2 deployed",
    "event.description": "Rollout of v1.4.2 containing performance patches for SQL queries.",
    "service.name": "order-service",
    "environment": "prod",
    "deployment.version": "v1.4.2",
    "deployment.commit": "7a3f81e",
    "actor.name": "GitHub Actions"
  }
}
```

---

## 5. Grafana Integration & Correlation
Correlation is handled natively by Grafana using standard data source links and LogQL queries, eliminating PostgreSQL dependencies.

### 5.1 Native Annotations (Event Visualization)
To draw vertical annotation lines on dashboards when deployments or incidents happen, Grafana queries Loki directly for operational logs matching the `event.name` attribute:

*   **Annotation Source:** Loki Data Source
*   **Query (LogQL):** `{environment="$environment"} | json | event_name="deployment" or event_name="incident"`
*   **Annotation Fields mapping:**
    *   *Title:* `attributes.event_title`
    *   *Text:* `attributes.event_description` (or `message`)
    *   *Tags:* `service_name, environment, attributes.event_status`

### 5.2 Trace-to-Log Correlation
Loki logs are linked to Tempo traces using derived fields in the Loki data source config.
*   **Regex Matcher:** `\"trace_id\":\"(\\w+)\"` or `trace_id=(\\w+)`
*   **Target Data Source:** Tempo

---

## 6. SRE Configuration Policies
### 6.1 Tail-Based Sampling
To optimize network and storage costs, the OTel Collector must be configured with tail-based sampling rules:
*   **Errors:** Keep 100% of traces where the status code is `Error` or contains exception logs.
*   **High Latency:** Keep 100% of traces with duration > `500ms`.
*   **Standard Traffic:** Keep a 1% random sample of healthy, low-latency traces.

### 6.2 Metrics Generation (Span Metrics Connector)
The OTel Collector dynamically reads spans passing through its pipelines and automatically computes RED metrics (Rate, Errors, Duration) for services, exporting them to Prometheus. This removes the need for manual metric instrumentations in Go code.

---

## 7. Developer Tooling: CLI Bootstrapper
To minimize configuration pain, a lightweight interactive CLI tool `deep-observer` is defined.

### 7.1 Command Surface
*   `deep-observer init`
    *   Initiates an interactive prompt in the target repository.
    *   Queries user for: Service Name, Language/Runtime, and Framework.
    *   Generates a helper module in the local repo (e.g., `pkg/telemetry/otel.go` for Go) containing clean initialization code.
    *   Generates pre-configured infrastructure files (`docker-compose.yaml`, `otel-collector.yaml`, `prometheus.yaml`, and `alert_rules.yaml`).
*   `deep-observer deploy --version <v> --commit <sha>`
    *   A helper command that packages and shoots a standard OTLP deployment event log to the configured collector endpoint. Useful for integration with CI/CD pipelines (e.g., GitHub Actions / GitLab CI).

---

## 8. Verification & Test Plan
*   **Functional Tests:** Verify that OTLP logs containing `event.name = "deployment"` successfully land in Loki and render as Annotations on Grafana dashboards.
*   **Correlation Tests:** Ensure clicking on a `trace_id` inside a Loki log entry correctly opens the corresponding Trace Graph in Grafana Tempo.
*   **Alerting Tests:** Induce artificial service errors and latency spikes using a test script to trigger Alertmanager notifications.
