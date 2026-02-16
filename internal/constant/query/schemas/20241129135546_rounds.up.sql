CREATE TABLE rounds(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status bet_status not null, 
    crash_poi DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP not null,
    closed_at TIMESTAMP
);