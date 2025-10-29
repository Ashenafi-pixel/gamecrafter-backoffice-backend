create table roll_da_dice (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status varchar not null DEFAULT 'ACTIVE',
    bet_amount decimal not null,
    won_status varchar ,
    crash_point decimal not null,
    timestamp timestamp not null,
    won_amount decimal,
    user_guessed_start_point decimal,
    user_guessed_end_point decimal,
    multiplier decimal,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)