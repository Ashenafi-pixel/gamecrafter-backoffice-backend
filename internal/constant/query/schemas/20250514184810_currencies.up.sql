CREATE TABLE currencies (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 name VARCHAR NOT NULL,
 Status VARCHAR NOT NULL DEFAULT 'ACTIVE',
 timestamp timestamp not null default now()
 );