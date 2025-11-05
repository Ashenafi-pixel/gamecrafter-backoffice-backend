-- Create alert email groups tables
-- This allows grouping emails together for alert notifications

-- Alert email groups table
CREATE TABLE IF NOT EXISTS alert_email_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    UNIQUE(name)
);

-- Alert email group members table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS alert_email_group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES alert_email_groups(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(group_id, email)
);

-- Add email_group_ids column to alert_configurations
-- Store as JSONB array of UUIDs
ALTER TABLE alert_configurations 
ADD COLUMN IF NOT EXISTS email_group_ids UUID[] DEFAULT '{}';

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_alert_email_groups_name ON alert_email_groups(name);
CREATE INDEX IF NOT EXISTS idx_alert_email_groups_created_by ON alert_email_groups(created_by);
CREATE INDEX IF NOT EXISTS idx_alert_email_group_members_group_id ON alert_email_group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_alert_email_group_members_email ON alert_email_group_members(email);
CREATE INDEX IF NOT EXISTS idx_alert_configurations_email_group_ids ON alert_configurations USING GIN(email_group_ids);

-- Create trigger to update updated_at
CREATE OR REPLACE FUNCTION update_alert_email_groups_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_alert_email_groups_updated_at
    BEFORE UPDATE ON alert_email_groups
    FOR EACH ROW
    EXECUTE FUNCTION update_alert_email_groups_updated_at();

