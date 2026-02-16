create table users_football_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status varchar not null DEFAULT 'PENDING',
    match_id UUID not null,
    selection VARCHAR not null DEFAULT 'DRAW',
    FOREIGN KEY (match_id) REFERENCES football_matchs(id) ON DELETE CASCADE
);

