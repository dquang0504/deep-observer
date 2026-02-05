# Deep Observer Troubleshooting Guide

This document records specific issues encountered during the development of Deep Observer and their solutions.

## 1. Observability Stack (Loki & Tempo)

### 1.1. Loki: `field enabled not found` (Pattern Ingester)
**Symptoms:**
- Loki container exits immediately.
- Logs (`docker logs deep-observer-loki`) show:
  ```
  failed parsing config: ... field x not found in type aggregation.Config ...
  ```
- Specifically related to `pattern_ingester`.

**Cause:**
The `pattern_ingester` feature is experimental and its configuration structure changes frequently between Loki versions. The configuration options (like `enabled` inside `metric_aggregation` or `lru_size`) were invalid for the running version.

**Fix:**
Remove or comment out the `pattern_ingester` block in `configs/loki/loki-config.yml`. It is not essential for basic logging functionality.

### 1.2. Tempo: `empty ring` Error
**Symptoms:**
- Tempo starts but queries fail with:
  ```
  error querying ingesters ... error getting replication set for ring (0): empty ring
  ```
- Happened when running Tempo in Monolithic mode (all-in-one).

**Cause:**
Tempo's Distributor and Ingester components need to know how to communicate (form a ring). In a standalone/docker-compose setup, they defaulted to an invalid state or couldn't find peers, leading to an empty ring.

**Fix:**
Downgrading Tempo to a stable version (e.g., **v2.9.1**) is the most reliable fix for this issue in monolithic mode. v2.10.0+ has known regressions with the ingester startup.

Also ensure `tempo.yml` is configured for in-memory:
```yaml
distributor:
  ring:
    kvstore:
      store: inmemory

ingester:
  lifecycler:
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
```

And ensuring the `command` in `docker-compose` includes `-target=all`.

### 1.3. Tempo: `field compactor not found`
**Symptoms:**
- Tempo container exits immediately.
- Logs show parsing error regarding `compactor`.

**Cause:**
In certain Tempo versions (or specfic monolithic configurations), the `compactor` block at the root level might be misplaced or its internal fields changed.

For MVP/Dev environments, simply comment out the `compactor` block in `configs/tempo/tempo.yml`. Default settings (or disabling it if not needed for short retention) are sufficient.

### 1.4. Tempo: `syntax error: unexpected IDENTIFIER` (TraceQL)
**Symptoms:**
- Searching in Grafana (e.g., Service Name = `sample-service`) fails with:
  `invalid TraceQL query: parse error ... syntax error: unexpected IDENTIFIER`

**Cause:**
This is a version mismatch between Grafana (frontend) and Tempo (backend). Newer Grafana versions generate TraceQL queries using *unquoted* identifiers (e.g., `{ name = sample-service }`), which older Tempo versions (like v2.6.0) cannot parse.

**Fix:**
- **Upgrade Tempo**: Use version **v2.8.0** or higher (Recommended: `v2.9.1`).
- **Workaround**: Manually quote string values in your TraceQL query: `{ name = "sample-service" }`.

### 1.5. Tempo: Missing Traces ("0 series returned")
**Symptoms:**
- Applications are running and logging trace generation.
- Queries in Grafana return "No traces" or "0 series returned".
- No errors in app logs.

**Cause:**
Tempo is running but not listening for OTLP traffic because the `receivers` configuration is missing.

**Fix:**
Ensure `distributor.receivers` is correctly defined in `configs/tempo/tempo.yml`:
```yaml
distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318
```
