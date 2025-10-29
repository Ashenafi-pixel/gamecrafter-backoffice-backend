create table if not exists loot_box_place_bets (
    id uuid not null primary key default gen_random_uuid(),
    user_id uuid not null,
    user_selection uuid null,
    loot_box jsonb not null,
    wonStatus varchar(10) not null default 'pending',
    status text not null default 'pending',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);