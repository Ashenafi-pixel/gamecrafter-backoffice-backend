-- Create game_import_config table for automated game import settings
CREATE TABLE IF NOT EXISTS public.game_import_config (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    brand_id uuid NOT NULL,
    schedule_type VARCHAR(20) NOT NULL CHECK (schedule_type IN ('daily', 'weekly', 'monthly', 'custom')),
    schedule_cron VARCHAR(100), -- For custom cron expressions
    providers JSONB, -- Array of provider names or null for all providers
    is_active BOOLEAN DEFAULT true,
    last_run_at TIMESTAMP WITH TIME ZONE,
    next_run_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT game_import_config_pkey PRIMARY KEY (id),
    CONSTRAINT fk_game_import_config_brand FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON DELETE CASCADE,
    CONSTRAINT uq_game_import_config_brand UNIQUE (brand_id)
);

COMMENT ON TABLE public.game_import_config IS 'Configuration for automated game import from Directus';
COMMENT ON COLUMN public.game_import_config.schedule_type IS 'Schedule type: daily, weekly, monthly, or custom';
COMMENT ON COLUMN public.game_import_config.schedule_cron IS 'Custom cron expression (only used when schedule_type is custom)';
COMMENT ON COLUMN public.game_import_config.providers IS 'JSON array of provider names to filter, or null for all providers';
COMMENT ON COLUMN public.game_import_config.is_active IS 'Whether the automated import is enabled';

CREATE INDEX IF NOT EXISTS idx_game_import_config_brand_id ON public.game_import_config(brand_id);
CREATE INDEX IF NOT EXISTS idx_game_import_config_is_active ON public.game_import_config(is_active);


