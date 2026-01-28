CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY,
    service_name TEXT NOT NULL UNIQUE,
    owner TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_services_name on services(service_name);

CREATE table IF NOT EXISTS environments(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE INDEX IF NOT EXISTS idx_environments_name ON environments(name);

CREATE TABLE IF NOT EXISTS deploy_events (
    id UUID PRIMARY KEY,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE RESTRICT,
    version TEXT NOT NULL,
    commit_hash TEXT,
    deployed_by TEXT,
    deployed_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_deploy_events_service ON deploy_events(service_id);
CREATE INDEX IF NOT EXISTS idx_deploy_events_env ON deploy_events(environment_id);
CREATE INDEX IF NOT EXISTS idx_deplpy_events_time ON deploy_events(deployed_at);
CREATE INDEX IF NOT EXISTS idx_deploy_events_service_time ON deploy_events(service_id, deployed_at);

CREATE TABLE IF NOT EXISTS retention_policies (
    id UUID PRIMARY KEY,
    signal_type TEXT NOT NULL UNIQUE CHECK (signal_type IN ('metrics', 'logs', 'traces')),
    retention_days INT NOT NULL CHECK (retention_days >= 0)
);

CREATE TABLE IF NOT EXISTS dashboards (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    grafana_uuid TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_dashboards_name ON dashboards(name);
CREATE INDEX IF NOT EXISTS idx_dashboards_uid ON dashboards(grafana_uuid);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL CHECK (event_type IN ('incident', 'config_change', 'maintenance')),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    severity TEXT,
    occurred_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_service ON events(service_id);
CREATE INDEX IF NOT EXISTS idx_events_environment ON events(environment_id);
CREATE INDEX IF NOT EXISTS idx_events_time ON events(occurred_at);
CREATE INDEX IF NOT EXISTS idx_events_type_time ON events(event_type, occurred_at);