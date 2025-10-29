CREATE TABLE IF NOT EXISTS agent_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR NOT NULL UNIQUE,
    client_secret TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    name TEXT NOT NULL,
    description TEXT,
    callback_url VARCHAR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
