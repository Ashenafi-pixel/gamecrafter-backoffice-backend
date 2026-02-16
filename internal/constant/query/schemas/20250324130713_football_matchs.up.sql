CREATE TABLE football_matchs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id UUID not null,
    league VARCHAR not null,
    date timestamp not null,
    home_team varchar not null,
    away_team varchar,
    status varchar DEFAULT 'ACTIVE',
    won varchar null,
    timestamp TIMESTAMP not null DEFAULT now(),
     FOREIGN KEY (round_id) REFERENCES football_match_rounds(id) ON DELETE CASCADE
 );