-- Manual KYC Tables Migration
-- Run this script manually on the database

-- 1. Update existing kyc_status values to new enum
UPDATE users SET kyc_status = 'NO_KYC' WHERE kyc_status = 'PENDING';
UPDATE users SET kyc_status = 'ID_VERIFIED' WHERE kyc_status = 'VERIFIED';

-- 2. Create KYC Documents table
CREATE TABLE IF NOT EXISTS kyc_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL, -- 'ID_FRONT', 'ID_BACK', 'PROOF_OF_ADDRESS', 'SELFIE_WITH_ID', 'BANK_STATEMENT', 'SOF_DOCUMENT'
    file_url TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    upload_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'PENDING', -- 'PENDING', 'APPROVED', 'REJECTED'
    rejection_reason TEXT,
    reviewed_by UUID REFERENCES users(id),
    review_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Create KYC Submissions table (tracks KYC application history)
CREATE TABLE IF NOT EXISTS kyc_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    submission_type VARCHAR(20) NOT NULL, -- 'INITIAL', 'RESUBMISSION', 'UPGRADE'
    status VARCHAR(20) NOT NULL, -- 'SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED', 'EXPIRED'
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    admin_notes TEXT,
    auto_triggered BOOLEAN DEFAULT FALSE,
    trigger_reason VARCHAR(100), -- 'DEPOSIT_THRESHOLD', 'WITHDRAWAL_THRESHOLD', 'MANUAL', 'TIME_BASED'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Create KYC Status Changes audit table
CREATE TABLE IF NOT EXISTS kyc_status_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    old_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL,
    changed_by UUID NOT NULL REFERENCES users(id),
    change_reason TEXT,
    admin_notes TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Create KYC Settings table for thresholds and rules
CREATE TABLE IF NOT EXISTS kyc_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value JSONB NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Insert default KYC settings
INSERT INTO kyc_settings (setting_key, setting_value, description) VALUES
('kyc_thresholds', '{"deposit_threshold_usd": 1000, "withdrawal_threshold_usd": 500, "daily_limit_usd": 5000}', 'KYC trigger thresholds'),
('document_requirements', '{"id_required": true, "proof_of_address_required": true, "selfie_required": true, "sof_required": false}', 'Required document types'),
('auto_kyc_rules', '{"enable_auto_triggers": true, "require_kyc_for_withdrawals": true, "kyc_expiry_days": 365}', 'Automatic KYC rules')
ON CONFLICT (setting_key) DO NOTHING;

-- 7. Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_kyc_documents_user_id ON kyc_documents(user_id);
CREATE INDEX IF NOT EXISTS idx_kyc_documents_status ON kyc_documents(status);
CREATE INDEX IF NOT EXISTS idx_kyc_documents_type ON kyc_documents(document_type);
CREATE INDEX IF NOT EXISTS idx_kyc_submissions_user_id ON kyc_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_kyc_submissions_status ON kyc_submissions(status);
CREATE INDEX IF NOT EXISTS idx_kyc_status_changes_user_id ON kyc_status_changes(user_id);
CREATE INDEX IF NOT EXISTS idx_users_kyc_status ON users(kyc_status);

-- 8. Add withdrawal restriction flag to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS withdrawal_restricted BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS kyc_required_for_withdrawal BOOLEAN DEFAULT TRUE;

-- 9. Create trigger to log KYC status changes
CREATE OR REPLACE FUNCTION log_kyc_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.kyc_status != NEW.kyc_status THEN
        INSERT INTO kyc_status_changes (user_id, old_status, new_status, changed_by, change_reason)
        VALUES (NEW.id, OLD.kyc_status, NEW.kyc_status, NEW.id, 'Status updated');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_kyc_status_change
    AFTER UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION log_kyc_status_change();

-- 10. Create function to check KYC requirements
CREATE OR REPLACE FUNCTION check_kyc_requirement(
    p_user_id UUID,
    p_transaction_type VARCHAR(20),
    p_amount_usd DECIMAL
) RETURNS BOOLEAN AS $$
DECLARE
    user_kyc_status VARCHAR(20);
    thresholds JSONB;
    threshold_value DECIMAL;
BEGIN
    -- Get user's current KYC status
    SELECT kyc_status INTO user_kyc_status
    FROM users WHERE id = p_user_id;
    
    -- Get KYC thresholds
    SELECT setting_value INTO thresholds
    FROM kyc_settings WHERE setting_key = 'kyc_thresholds';
    
    -- Check if KYC is required based on transaction type and amount
    IF p_transaction_type = 'DEPOSIT' THEN
        threshold_value := (thresholds->>'deposit_threshold_usd')::DECIMAL;
    ELSIF p_transaction_type = 'WITHDRAWAL' THEN
        threshold_value := (thresholds->>'withdrawal_threshold_usd')::DECIMAL;
    ELSE
        RETURN TRUE; -- No KYC required for other transaction types
    END IF;
    
    -- If amount exceeds threshold and user doesn't have sufficient KYC, return FALSE
    IF p_amount_usd >= threshold_value AND user_kyc_status NOT IN ('ID_VERIFIED', 'ID_SOF_VERIFIED') THEN
        RETURN FALSE;
    END IF;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- 11. Add comments for documentation
COMMENT ON TABLE kyc_documents IS 'Stores uploaded KYC documents with review status';
COMMENT ON TABLE kyc_submissions IS 'Tracks KYC application submissions and their status';
COMMENT ON TABLE kyc_status_changes IS 'Audit trail for KYC status changes';
COMMENT ON TABLE kyc_settings IS 'Configurable KYC rules and thresholds';
COMMENT ON FUNCTION check_kyc_requirement IS 'Checks if user meets KYC requirements for transaction amount';
