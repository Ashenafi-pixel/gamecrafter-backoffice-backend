-- Migration: Add Allowed Pages System
-- Description: Creates pages table and user_allowed_pages junction table for page-based access control

-- Create pages table
CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path VARCHAR(255) NOT NULL UNIQUE,
    label VARCHAR(255) NOT NULL,
    parent_id UUID NULL,
    icon VARCHAR(100) NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_pages_parent FOREIGN KEY (parent_id) REFERENCES pages(id) ON DELETE CASCADE
);

-- Create user_allowed_pages junction table
CREATE TABLE IF NOT EXISTS user_allowed_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    page_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user_allowed_pages_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_allowed_pages_page FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE,
    CONSTRAINT uq_user_allowed_pages UNIQUE (user_id, page_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_pages_parent_id ON pages(parent_id);
CREATE INDEX IF NOT EXISTS idx_pages_path ON pages(path);
CREATE INDEX IF NOT EXISTS idx_user_allowed_pages_user_id ON user_allowed_pages(user_id);
CREATE INDEX IF NOT EXISTS idx_user_allowed_pages_page_id ON user_allowed_pages(page_id);

-- Add comments for documentation
COMMENT ON TABLE pages IS 'Stores all available pages/routes in the system';
COMMENT ON TABLE user_allowed_pages IS 'Junction table mapping users to their allowed pages';
COMMENT ON COLUMN pages.parent_id IS 'Reference to parent page (for hierarchical structure - sidebar items as parents)';
COMMENT ON COLUMN pages.path IS 'Unique route path (e.g., /dashboard, /players)';
COMMENT ON COLUMN pages.label IS 'Display name for the page (e.g., Dashboard, Player Management)';

