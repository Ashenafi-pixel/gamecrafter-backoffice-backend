CREATE TABLE leagues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    league_name varchar not null,
    status varchar DEFAULT 'ACTIVE',
    timestamp TIMESTAMP not null DEFAULT now()
 );