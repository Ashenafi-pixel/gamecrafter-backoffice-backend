-- Migration: Create global_rakeback_override table
-- This allows administrators to temporarily override all VIP rakeback levels with a single global percentage
-- Used for marketing events like "Happy Hours"

CREATE TABLE global_rakeback_override (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_active BOOLEAN DEFAULT false,
    rakeback_percentage DECIMAL(5,2) DEFAULT 0.00, -- Up to 2 decimal places (e.g., 100.00, 12.5)
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for active override queries
CREATE INDEX IF NOT EXISTS idx_global_rakeback_override_active ON global_rakeback_override(is_active) WHERE is_active = true;

-- Add comment to explain the table
COMMENT ON TABLE global_rakeback_override IS 'Global rakeback override for marketing events. When active, overrides all VIP-level rakeback percentages with a single global value.';

