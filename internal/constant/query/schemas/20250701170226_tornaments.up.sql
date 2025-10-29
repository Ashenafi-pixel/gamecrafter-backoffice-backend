create table tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rank text not null,
    level int not null,
    cumulative_points int not null,
    rewards jsonb not null,
    created_at timestamptz default now() not null,
    updated_at timestamptz default now() not null,
    deleted_at timestamptz
);

create index idx_tournaments_rank on tournaments (rank);
