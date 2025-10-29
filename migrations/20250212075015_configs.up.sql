CREATE TABLE configs (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 name VARCHAR(20) Unique  Not Null,
 value VARCHAR not null,
 created_at TIMESTAMP not null DEFAULT now());