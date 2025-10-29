-- Add created_by column to user_notifications table
ALTER TABLE user_notifications ADD COLUMN created_by UUID REFERENCES users(id);
