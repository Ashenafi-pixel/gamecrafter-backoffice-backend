create table squads_earns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    squad_id UUID NOT NULL,
    user_id UUID NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'P',
    earned decimal NOT NULL DEFAULT 0,
    game_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_squad_earns_squad FOREIGN KEY (squad_id) REFERENCES squads(id) ON DELETE CASCADE,
    CONSTRAINT fk_squad_earns_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);