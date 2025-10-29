create table levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level decimal not null,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    created_by UUID NOT NULL,
    CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);


CREATE UNIQUE INDEX uniq_levels_active
    ON levels(level)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uniq_levels_deleted
    ON levels(level, deleted_at);