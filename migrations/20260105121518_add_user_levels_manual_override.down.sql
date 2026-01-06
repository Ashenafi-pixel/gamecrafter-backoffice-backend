-- Rollback: Remove manual override columns from user_levels table

-- Remove foreign key constraints first
ALTER TABLE public.user_levels 
DROP CONSTRAINT IF EXISTS user_levels_manual_override_tier_id_fkey;

ALTER TABLE public.user_levels 
DROP CONSTRAINT IF EXISTS user_levels_manual_override_set_by_fkey;

-- Remove manual override columns from user_levels table
ALTER TABLE public.user_levels 
DROP COLUMN IF EXISTS manual_override_tier_id;

ALTER TABLE public.user_levels 
DROP COLUMN IF EXISTS manual_override_set_by;

ALTER TABLE public.user_levels 
DROP COLUMN IF EXISTS manual_override_set_at;

ALTER TABLE public.user_levels 
DROP COLUMN IF EXISTS manual_override_level;

ALTER TABLE public.user_levels 
DROP COLUMN IF EXISTS is_manual_override;



