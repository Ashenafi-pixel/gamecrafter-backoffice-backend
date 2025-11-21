ALTER TABLE user_levels
    ADD COLUMN is_manual_override BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN manual_override_level INTEGER,
    ADD COLUMN manual_override_tier_id UUID REFERENCES cashback_tiers(id),
    ADD COLUMN manual_override_set_by UUID REFERENCES users(id),
    ADD COLUMN manual_override_set_at TIMESTAMPTZ;

