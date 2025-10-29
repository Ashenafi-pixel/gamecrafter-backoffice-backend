create table street_kings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status varchar not null DEFAULT 'ACTIVE',
    version varchar(50) not null,
    bet_amount decimal not null,
    won_amount decimal,
    crash_point decimal not null,
    cash_out_point decimal,
    timestamp timestamp not null,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );

