create table waiting_squad_members (
    id uuid primary key,
    user_id uuid not null references users(id) on delete cascade,
    squad_id uuid not null references squads(id) on delete cascade,
    created_at timestamptz default now() not null
);

-- Add unique constraint to prevent duplicate entries for the same user and squad
create unique index idx_waiting_squad_members_user_squad on waiting_squad_members (user_id, squad_id);