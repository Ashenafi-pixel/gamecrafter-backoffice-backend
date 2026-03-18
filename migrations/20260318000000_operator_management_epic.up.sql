-- Operator management tables (CRUD is based on existing operators table)
-- Adds: operator_credentials, operator_allowed_origins, operator_feature_flags, operator_providers
-- Also ensures operator_games exists (if not already created manually).

-- operator_credentials: API credentials per operator (api_key + signing_key)
CREATE TABLE IF NOT EXISTS operator_credentials (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(operator_id) ON DELETE CASCADE,
    api_key VARCHAR(255) NOT NULL,
    signing_key TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_operator_credentials_operator_id ON operator_credentials(operator_id);
CREATE INDEX IF NOT EXISTS idx_operator_credentials_api_key ON operator_credentials(api_key);

-- operator_allowed_origins: per-operator allowed origins (embed/cors)
CREATE TABLE IF NOT EXISTS operator_allowed_origins (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(operator_id) ON DELETE CASCADE,
    origin VARCHAR(500) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(operator_id, origin)
);

CREATE INDEX IF NOT EXISTS idx_operator_allowed_origins_operator_id ON operator_allowed_origins(operator_id);

-- operator_feature_flags: per-operator feature flags as JSONB (key -> enabled)
CREATE TABLE IF NOT EXISTS operator_feature_flags (
    operator_id INT PRIMARY KEY REFERENCES operators(operator_id) ON DELETE CASCADE,
    flags JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_operator_feature_flags_operator_id ON operator_feature_flags(operator_id);

-- operator_providers: assign providers to operators (many-to-many)
CREATE TABLE IF NOT EXISTS operator_providers (
    operator_id INT NOT NULL REFERENCES operators(operator_id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES game_providers(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (operator_id, provider_id)
);

CREATE INDEX IF NOT EXISTS idx_operator_providers_operator_id ON operator_providers(operator_id);
CREATE INDEX IF NOT EXISTS idx_operator_providers_provider_id ON operator_providers(provider_id);

-- operator_games: assign games to operators (many-to-many)
CREATE TABLE IF NOT EXISTS operator_games (
    operator_id INTEGER NOT NULL REFERENCES operators(operator_id) ON DELETE CASCADE,
    game_id     UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (operator_id, game_id)
);

CREATE INDEX IF NOT EXISTS idx_operator_games_operator_id ON operator_games(operator_id);
CREATE INDEX IF NOT EXISTS idx_operator_games_game_id ON operator_games(game_id);

