-- Admin Activity Logs Migration
-- This migration creates a comprehensive admin activity logging system

-- Create admin_activity_logs table
CREATE TABLE admin_activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL, -- e.g., 'user_blocked', 'balance_updated', 'withdrawal_approved'
    resource_type VARCHAR(50) NOT NULL, -- e.g., 'user', 'withdrawal', 'balance', 'system_config'
    resource_id UUID, -- ID of the affected resource (user, withdrawal, etc.)
    description TEXT NOT NULL, -- Human-readable description of the action
    details JSONB, -- Additional structured data about the action
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    severity VARCHAR(20) DEFAULT 'info' CHECK (severity IN ('low', 'info', 'warning', 'error', 'critical')),
    category VARCHAR(50) NOT NULL, -- e.g., 'user_management', 'financial', 'security', 'system'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_admin_activity_logs_admin_user_id ON admin_activity_logs(admin_user_id);
CREATE INDEX idx_admin_activity_logs_action ON admin_activity_logs(action);
CREATE INDEX idx_admin_activity_logs_resource_type ON admin_activity_logs(resource_type);
CREATE INDEX idx_admin_activity_logs_resource_id ON admin_activity_logs(resource_id);
CREATE INDEX idx_admin_activity_logs_category ON admin_activity_logs(category);
CREATE INDEX idx_admin_activity_logs_severity ON admin_activity_logs(severity);
CREATE INDEX idx_admin_activity_logs_created_at ON admin_activity_logs(created_at);
CREATE INDEX idx_admin_activity_logs_composite ON admin_activity_logs(admin_user_id, created_at DESC);

-- Create admin_activity_categories table for predefined categories
CREATE TABLE admin_activity_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    color VARCHAR(7), -- Hex color code for UI
    icon VARCHAR(50), -- Icon name for UI
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default categories
INSERT INTO admin_activity_categories (name, description, color, icon) VALUES
('user_management', 'User account management activities', '#3B82F6', 'users'),
('financial', 'Financial transactions and balance management', '#10B981', 'dollar-sign'),
('security', 'Security-related actions and access control', '#EF4444', 'shield'),
('system', 'System configuration and maintenance', '#8B5CF6', 'settings'),
('withdrawal', 'Withdrawal management and processing', '#F59E0B', 'arrow-up'),
('game_management', 'Game configuration and management', '#EC4899', 'gamepad'),
('reports', 'Report generation and analytics', '#06B6D4', 'chart-bar'),
('notifications', 'Notification and communication management', '#84CC16', 'bell');

-- Create admin_activity_actions table for predefined actions
CREATE TABLE admin_activity_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    category_id UUID REFERENCES admin_activity_categories(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert common admin actions
INSERT INTO admin_activity_actions (name, description, category_id) VALUES
-- User Management Actions
('user_created', 'User account created', (SELECT id FROM admin_activity_categories WHERE name = 'user_management')),
('user_updated', 'User account updated', (SELECT id FROM admin_activity_categories WHERE name = 'user_management')),
('user_blocked', 'User account blocked', (SELECT id FROM admin_activity_categories WHERE name = 'user_management')),
('user_unblocked', 'User account unblocked', (SELECT id FROM admin_activity_categories WHERE name = 'user_management')),
('user_deleted', 'User account deleted', (SELECT id FROM admin_activity_categories WHERE name = 'user_management')),
('user_verified', 'User account verified', (SELECT id FROM admin_activity_categories WHERE name = 'user_management')),
('user_kyc_approved', 'User KYC approved', (SELECT id FROM admin_activity_categories WHERE name = 'user_management')),
('user_kyc_rejected', 'User KYC rejected', (SELECT id FROM admin_activity_categories WHERE name = 'user_management')),

-- Financial Actions
('balance_credited', 'User balance credited', (SELECT id FROM admin_activity_categories WHERE name = 'financial')),
('balance_debited', 'User balance debited', (SELECT id FROM admin_activity_categories WHERE name = 'financial')),
('manual_fund_added', 'Manual funds added to user account', (SELECT id FROM admin_activity_categories WHERE name = 'financial')),
('manual_fund_removed', 'Manual funds removed from user account', (SELECT id FROM admin_activity_categories WHERE name = 'financial')),
('refund_processed', 'Refund processed for user', (SELECT id FROM admin_activity_categories WHERE name = 'financial')),

-- Withdrawal Actions
('withdrawal_approved', 'Withdrawal request approved', (SELECT id FROM admin_activity_categories WHERE name = 'withdrawal')),
('withdrawal_rejected', 'Withdrawal request rejected', (SELECT id FROM admin_activity_categories WHERE name = 'withdrawal')),
('withdrawal_paused', 'Withdrawal request paused', (SELECT id FROM admin_activity_categories WHERE name = 'withdrawal')),
('withdrawal_unpaused', 'Withdrawal request unpaused', (SELECT id FROM admin_activity_categories WHERE name = 'withdrawal')),
('withdrawal_global_pause', 'Global withdrawal pause toggled', (SELECT id FROM admin_activity_categories WHERE name = 'withdrawal')),

-- System Actions
('system_config_updated', 'System configuration updated', (SELECT id FROM admin_activity_categories WHERE name = 'system')),
('2fa_enabled', '2FA enabled for user', (SELECT id FROM admin_activity_categories WHERE name = 'security')),
('2fa_disabled', '2FA disabled for user', (SELECT id FROM admin_activity_categories WHERE name = 'security')),
('password_reset', 'Password reset for user', (SELECT id FROM admin_activity_categories WHERE name = 'security')),
('login_successful', 'Admin login successful', (SELECT id FROM admin_activity_categories WHERE name = 'security')),
('login_failed', 'Admin login failed', (SELECT id FROM admin_activity_categories WHERE name = 'security'));

-- Create function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_admin_activity_logs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for updated_at
CREATE TRIGGER trigger_update_admin_activity_logs_updated_at
    BEFORE UPDATE ON admin_activity_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_activity_logs_updated_at();
