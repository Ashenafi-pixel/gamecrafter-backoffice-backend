-- Global Rakeback Override Migration (Happy Hour Mode)
-- This migration adds the global rakeback override feature that allows admins
-- to temporarily override all VIP-based rakeback percentages with a single
-- global percentage for promotional events like "Happy Hours"

-- Create global_rakeback_override table
CREATE TABLE IF NOT EXISTS global_rakeback_override (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_enabled BOOLEAN NOT NULL DEFAULT false,
    override_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    enabled_by UUID REFERENCES users(id) ON DELETE SET NULL,
    enabled_at TIMESTAMP WITH TIME ZONE,
    disabled_by UUID REFERENCES users(id) ON DELETE SET NULL,
    disabled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT single_config CHECK (id = '00000000-0000-0000-0000-000000000001'::uuid),
    CONSTRAINT valid_percentage CHECK (override_percentage >= 0.00 AND override_percentage <= 100.00)
);

-- Create index for quick lookup on is_enabled
CREATE INDEX IF NOT EXISTS idx_global_rakeback_override_enabled 
ON global_rakeback_override(is_enabled);

-- Insert default configuration (disabled by default)
INSERT INTO global_rakeback_override (id, is_enabled, override_percentage) 
VALUES ('00000000-0000-0000-0000-000000000001'::uuid, false, 0.00)
ON CONFLICT (id) DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE global_rakeback_override IS 
'Global rakeback override configuration for Happy Hour promotions. Only one row should exist (singleton pattern).';

COMMENT ON COLUMN global_rakeback_override.is_enabled IS 
'When true, override_percentage applies to all users regardless of VIP tier';

COMMENT ON COLUMN global_rakeback_override.override_percentage IS 
'Global rakeback percentage (0.00-100.00) that overrides VIP tier rates when enabled';

COMMENT ON COLUMN global_rakeback_override.enabled_by IS 
'Admin user ID who activated the override';

COMMENT ON COLUMN global_rakeback_override.enabled_at IS 
'Timestamp when the override was activated';

COMMENT ON COLUMN global_rakeback_override.disabled_by IS 
'Admin user ID who deactivated the override';

COMMENT ON COLUMN global_rakeback_override.disabled_at IS 
'Timestamp when the override was deactivated';

