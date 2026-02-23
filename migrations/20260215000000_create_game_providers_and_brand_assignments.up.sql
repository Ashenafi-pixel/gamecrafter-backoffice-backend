-- Game providers table (EPIC: Providers & Games)
CREATE TABLE IF NOT EXISTS game_providers (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    name character varying(255) NOT NULL,
    code character varying(100) NOT NULL UNIQUE,
    status character varying(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'MAINTENANCE')),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_game_providers_code ON game_providers(code);
CREATE INDEX idx_game_providers_status ON game_providers(status);

COMMENT ON TABLE game_providers IS 'Game content providers (e.g. Pragmatic Play, Evolution)';
COMMENT ON COLUMN game_providers.code IS 'Unique short code for the provider';

-- Brand-Provider assignment (many-to-many)
CREATE TABLE IF NOT EXISTS brand_providers (
    brand_id uuid NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    provider_id uuid NOT NULL REFERENCES game_providers(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (brand_id, provider_id)
);

CREATE INDEX idx_brand_providers_brand_id ON brand_providers(brand_id);
CREATE INDEX idx_brand_providers_provider_id ON brand_providers(provider_id);

COMMENT ON TABLE brand_providers IS 'Assigns which providers are available per brand';

-- Brand-Game assignment (many-to-many)
CREATE TABLE IF NOT EXISTS brand_games (
    brand_id uuid NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    game_id uuid NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (brand_id, game_id)
);

CREATE INDEX idx_brand_games_brand_id ON brand_games(brand_id);
CREATE INDEX idx_brand_games_game_id ON brand_games(game_id);

COMMENT ON TABLE brand_games IS 'Assigns which games are available per brand';
