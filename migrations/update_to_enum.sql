-- Update existing tables to use notification_type enum
-- Run these queries after the tables have been created

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

-- 2. Update message_campaigns table to use enum
-- First, add a temporary column with the enum type
ALTER TABLE message_campaigns ADD COLUMN message_type_new notification_type;

-- Update the new column with data from the old column
UPDATE message_campaigns SET message_type_new = message_type::notification_type;

-- Drop the old column and rename the new one
ALTER TABLE message_campaigns DROP COLUMN message_type;
ALTER TABLE message_campaigns RENAME COLUMN message_type_new TO message_type;

-- Make it NOT NULL
ALTER TABLE message_campaigns ALTER COLUMN message_type SET NOT NULL;

-- 3. Update user_notifications table to use enum (if it exists and uses TEXT type)
-- First check if the type column exists and is TEXT
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'user_notifications' AND column_name = 'type' AND data_type = 'text') THEN
        
        -- Add a temporary column with the enum type
        ALTER TABLE user_notifications ADD COLUMN type_new notification_type;
        
        -- Update the new column with data from the old column
        UPDATE user_notifications SET type_new = type::notification_type;
        
        -- Drop the old column and rename the new one
        ALTER TABLE user_notifications DROP COLUMN type;
        ALTER TABLE user_notifications RENAME COLUMN type_new TO type;
        
        -- Make it NOT NULL
        ALTER TABLE user_notifications ALTER COLUMN type SET NOT NULL;
    END IF;
END $$;
