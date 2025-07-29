create table users_football_matche_rounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status varchar not null DEFAULT 'ACTIVE',
    won_status varchar DEFAULT 'PENDING',
    user_id UUID not null,
    football_round_id UUID, 
    bet_amount decimal,
    won_amount decimal not null,
    timestamp timestamp,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (football_round_id) REFERENCES football_match_rounds(id) ON DELETE CASCADE
);

