create table squads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    handle varchar not null,
    owner UUID not null,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_squad_owner FOREIGN KEY (owner) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX uniq_squads_active
    ON squads(handle)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uniq_squads_deleted
    ON squads(handle, deleted_at);