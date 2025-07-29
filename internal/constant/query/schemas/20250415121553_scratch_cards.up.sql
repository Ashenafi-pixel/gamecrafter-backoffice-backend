create table scratch_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status varchar not null DEFAULT 'ACTIVE',
    bet_amount decimal not null,
    won_status varchar ,
    timestamp timestamp not null,
    won_amount decimal,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)