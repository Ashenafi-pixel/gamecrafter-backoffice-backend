create table lottery_logs (
    id uuid primary key default gen_random_uuid(),
    lottery_id uuid not null references lotteries(id) on delete cascade,
    lottery_reward_id uuid not null,
    draw_numbers varchar[] not null,
    prize decimal not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);