-- Enhanced Notification System Migration
-- This migration adds support for message campaigns, user segmentation, and enhanced notification types

-- 1. Create notification_type enum if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_type') THEN
        CREATE TYPE notification_type AS ENUM (
            'Promotional', 'KYC', 'Bonus', 'Welcome', 'System', 'Alert',
            'payments', 'security', 'general'
        );
    END IF;
END $$;

-- 2. Create message_campaigns table
CREATE TABLE message_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    message_type notification_type NOT NULL,
    subject VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'cancelled')),
    scheduled_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    total_recipients INTEGER DEFAULT 0,
    delivered_count INTEGER DEFAULT 0,
    read_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Create message_segments table
CREATE TABLE message_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES message_campaigns(id) ON DELETE CASCADE,
    segment_type VARCHAR(50) NOT NULL CHECK (segment_type IN ('criteria', 'csv', 'all_users')),
    segment_name VARCHAR(255),
    criteria JSONB, -- For criteria-based segmentation
    csv_data TEXT, -- For CSV upload data
    user_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Create campaign_recipients table
CREATE TABLE campaign_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES message_campaigns(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_id UUID REFERENCES user_notifications(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'delivered', 'read', 'failed')),
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(campaign_id, user_id)
);

-- 5. Create user_activity_log table for segmentation
CREATE TABLE user_activity_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL,
    activity_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Add indexes for performance
CREATE INDEX idx_message_campaigns_created_by ON message_campaigns(created_by);
CREATE INDEX idx_message_campaigns_status ON message_campaigns(status);
CREATE INDEX idx_message_campaigns_scheduled_at ON message_campaigns(scheduled_at);
CREATE INDEX idx_message_campaigns_message_type ON message_campaigns(message_type);

CREATE INDEX idx_message_segments_campaign_id ON message_segments(campaign_id);
CREATE INDEX idx_message_segments_segment_type ON message_segments(segment_type);

CREATE INDEX idx_campaign_recipients_campaign_id ON campaign_recipients(campaign_id);
CREATE INDEX idx_campaign_recipients_user_id ON campaign_recipients(user_id);
CREATE INDEX idx_campaign_recipients_status ON campaign_recipients(status);
CREATE INDEX idx_campaign_recipients_notification_id ON campaign_recipients(notification_id);

CREATE INDEX idx_user_activity_log_user_id ON user_activity_log(user_id);
CREATE INDEX idx_user_activity_log_activity_type ON user_activity_log(activity_type);
CREATE INDEX idx_user_activity_log_created_at ON user_activity_log(created_at);

-- 7. Add created_by column to user_notifications if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'user_notifications' AND column_name = 'created_by') THEN
        ALTER TABLE user_notifications ADD COLUMN created_by UUID REFERENCES users(id);
    END IF;
END $$;

-- 8. Add updated_at trigger for message_campaigns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_message_campaigns_updated_at 
    BEFORE UPDATE ON message_campaigns 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
