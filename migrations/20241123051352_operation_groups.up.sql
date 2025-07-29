create table operational_groups(
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name varchar(50),
     description text, 
     created_at TIMESTAMP
);