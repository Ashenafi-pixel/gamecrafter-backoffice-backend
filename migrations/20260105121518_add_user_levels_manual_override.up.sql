-- Migration: Add manual override columns to user_levels table
-- These columns allow admins to manually override a user's level/tier
-- The columns already exist on dev, so we use IF NOT EXISTS for safety

-- Add manual override columns to user_levels table
ALTER TABLE public.user_levels 
ADD COLUMN IF NOT EXISTS is_manual_override bool NOT NULL DEFAULT false;

ALTER TABLE public.user_levels 
ADD COLUMN IF NOT EXISTS manual_override_level int4;

ALTER TABLE public.user_levels 
ADD COLUMN IF NOT EXISTS manual_override_set_at timestamptz;

ALTER TABLE public.user_levels 
ADD COLUMN IF NOT EXISTS manual_override_set_by uuid;

ALTER TABLE public.user_levels 
ADD COLUMN IF NOT EXISTS manual_override_tier_id uuid;

-- Add comments for documentation
COMMENT ON COLUMN public.user_levels.is_manual_override IS 'Indicates if the user level has been manually overridden by an admin';
COMMENT ON COLUMN public.user_levels.manual_override_level IS 'Manually set level when override is active';
COMMENT ON COLUMN public.user_levels.manual_override_set_at IS 'Timestamp when the manual override was set';
COMMENT ON COLUMN public.user_levels.manual_override_set_by IS 'UUID of the admin who set the manual override';
COMMENT ON COLUMN public.user_levels.manual_override_tier_id IS 'UUID of the manually assigned tier when override is active';

-- Add foreign key constraints if they don't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'user_levels_manual_override_set_by_fkey'
    ) THEN
        ALTER TABLE public.user_levels 
        ADD CONSTRAINT user_levels_manual_override_set_by_fkey 
        FOREIGN KEY (manual_override_set_by) REFERENCES public.users(id) ON DELETE SET NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'user_levels_manual_override_tier_id_fkey'
    ) THEN
        ALTER TABLE public.user_levels 
        ADD CONSTRAINT user_levels_manual_override_tier_id_fkey 
        FOREIGN KEY (manual_override_tier_id) REFERENCES public.cashback_tiers(id) ON DELETE SET NULL;
    END IF;
END $$;

