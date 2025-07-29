CREATE TABLE users (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 username VARCHAR(20) Unique  Not Null,
 phone_number VARCHAR(15) Unique Not Null,
 password TEXT not null,
 created_at TIMESTAMP not null DEFAULT now()
);