CREATE TABLE IF NOT EXISTS adds_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_url VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    service_id VARCHAR(255) NOT NULL,
    service_secret VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(255) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE UNIQUE INDEX uniq_adds_services_service_id_active
    ON adds_services(service_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uniq_adds_services_service_id_deleted
    ON adds_services(service_id, deleted_at);

CREATE INDEX idx_adds_services_status ON adds_services(status);
CREATE INDEX idx_adds_services_created_at ON adds_services(created_at);
