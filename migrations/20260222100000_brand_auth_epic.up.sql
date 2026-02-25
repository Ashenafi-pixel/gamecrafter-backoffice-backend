
-- brand_credentials: API credentials per brand (client_id + hashed secret; optional signing_key for HMAC)
CREATE TABLE IF NOT EXISTS brand_credentials (
    id SERIAL PRIMARY KEY,
    brand_id INT NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret_hash TEXT NOT NULL,
    signing_key_encrypted TEXT,
    name VARCHAR(100) NOT NULL DEFAULT 'Default',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_rotated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(brand_id, name)
);

CREATE INDEX IF NOT EXISTS idx_brand_credentials_brand_id ON brand_credentials(brand_id);
CREATE INDEX IF NOT EXISTS idx_brand_credentials_client_id ON brand_credentials(client_id);

-- brand_allowed_origins: per-brand CORS / allowed origins
CREATE TABLE IF NOT EXISTS brand_allowed_origins (
    id SERIAL PRIMARY KEY,
    brand_id INT NOT NULL,
    origin VARCHAR(500) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(brand_id, origin)
);

CREATE INDEX IF NOT EXISTS idx_brand_allowed_origins_brand_id ON brand_allowed_origins(brand_id);

-- brand_feature_flags: per-brand feature flags (key -> enabled)
CREATE TABLE IF NOT EXISTS brand_feature_flags (
    brand_id INT NOT NULL,
    flag_key VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (brand_id, flag_key)
);

CREATE INDEX IF NOT EXISTS idx_brand_feature_flags_brand_id ON brand_feature_flags(brand_id);
