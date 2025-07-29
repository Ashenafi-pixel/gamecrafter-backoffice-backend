CREATE TABLE football_match_rounds(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status varchar,
    timestamp TIMESTAMP not null DEFAULT now()
 );