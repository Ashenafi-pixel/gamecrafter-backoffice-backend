CREATE TABLE clubs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_name varchar not null,
    status varchar DEFAULT 'ACTIVE',
    timestamp TIMESTAMP not null DEFAULT now()
 );