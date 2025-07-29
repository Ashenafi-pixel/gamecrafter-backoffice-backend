create table operational_types (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     group_id UUID not null,
     name varchar (50),
     description text,
     created_at TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES operational_groups(id) ON DELETE CASCADE
);