create table lotteries (
    id uuid  primary key default gen_random_uuid(),
    name text not null,
    price decimal not null,
    min_selectable integer not null,
    max_selectable integer not null,
    draw_frequency text not null,
    number_of_balls int not null,
    description text not null,
    status text not null default 'active',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
