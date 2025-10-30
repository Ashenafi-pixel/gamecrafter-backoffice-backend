-- Update existing kyc_documents table to match backend expectations
-- Run this to fix the column mismatch

-- Add file_url column if it doesn't exist
ALTER TABLE kyc_documents ADD COLUMN IF NOT EXISTS file_url TEXT;

-- Update existing records to copy file_path to file_url (if file_path exists)
UPDATE kyc_documents SET file_url = file_path WHERE file_url IS NULL AND file_path IS NOT NULL;

-- Make file_url NOT NULL (this will fail if there are NULL values)
ALTER TABLE kyc_documents ALTER COLUMN file_url SET NOT NULL;

-- Drop the old file_path column if it exists
ALTER TABLE kyc_documents DROP COLUMN IF EXISTS file_path;

-- Drop other columns that aren't used by the backend
ALTER TABLE kyc_documents DROP COLUMN IF EXISTS file_size;
ALTER TABLE kyc_documents DROP COLUMN IF EXISTS mime_type;

-- Rename reviewed_at to review_date to match backend expectations
ALTER TABLE kyc_documents RENAME COLUMN reviewed_at TO review_date;

-- Verify the table structure
\d kyc_documents;
