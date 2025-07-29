create table lottery_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id varchar NOT NULL,
    client_secret TEXT NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'active',
    name TEXT NOT NULL,
    description TEXT,
    callback_url varchar,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uniq_lottery_client_id_active
    ON lottery_services(client_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uniq_lottery_client_id_deleted
    ON lottery_services(client_id, deleted_at);