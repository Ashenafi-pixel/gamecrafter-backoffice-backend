create table temp (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now(),
    data jsonb not null
);