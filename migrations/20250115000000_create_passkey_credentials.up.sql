-- Create passkey_credentials table for storing WebAuthn passkey data
CREATE TABLE IF NOT EXISTS passkey_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id TEXT NOT NULL,
    raw_id BYTEA NOT NULL,
    public_key BYTEA NOT NULL,
    attestation_object BYTEA NOT NULL,
    client_data_json BYTEA NOT NULL,
    counter BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT 'Passkey',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    UNIQUE(user_id, credential_id)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_passkey_credentials_user_id ON passkey_credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_passkey_credentials_credential_id ON passkey_credentials(credential_id);
CREATE INDEX IF NOT EXISTS idx_passkey_credentials_active ON passkey_credentials(user_id, is_active) WHERE is_active = true;

-- Add comment
COMMENT ON TABLE passkey_credentials IS 'Stores WebAuthn passkey credentials for 2FA authentication';
COMMENT ON COLUMN passkey_credentials.credential_id IS 'Base64-encoded credential ID from WebAuthn';
COMMENT ON COLUMN passkey_credentials.raw_id IS 'Raw credential ID bytes';
COMMENT ON COLUMN passkey_credentials.public_key IS 'Public key from WebAuthn attestation';
COMMENT ON COLUMN passkey_credentials.attestation_object IS 'WebAuthn attestation object';
COMMENT ON COLUMN passkey_credentials.client_data_json IS 'WebAuthn client data JSON';
COMMENT ON COLUMN passkey_credentials.counter IS 'WebAuthn signature counter';
COMMENT ON COLUMN passkey_credentials.name IS 'User-friendly name for the passkey';
COMMENT ON COLUMN passkey_credentials.last_used_at IS 'When the passkey was last used for authentication';
COMMENT ON COLUMN passkey_credentials.is_active IS 'Whether the passkey is active and can be used';
