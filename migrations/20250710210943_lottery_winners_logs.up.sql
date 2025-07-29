create table lottery_winners_logs (
    id uuid primary key default gen_random_uuid(),
    lottery_id uuid not null references lotteries(id) on delete cascade,
    user_id uuid not null references users(id) on delete cascade,
    reward_id uuid not null ,
    won_amount decimal not null,
    currency text not null default 'P',
    number_of_tickets int not null,
    ticket_number text not null,
    status text not null default 'closed',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);