create table spinning_wheels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status varchar not null DEFAULT 'ACTIVE',
    bet_amount varchar not null,
    timestamp timestamp not null,
    won_amount varchar,
    won_status varchar,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)