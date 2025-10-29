create table if not exists loot_box (
    id uuid not null primary key default gen_random_uuid(),
    type text not null,
    prizeAmount decimal not null,
    weight decimal not null, 
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);