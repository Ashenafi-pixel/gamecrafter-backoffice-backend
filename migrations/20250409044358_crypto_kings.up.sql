create table crypto_kings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status varchar not null DEFAULT 'ACTIVE',
    bet_amount decimal not null,
    won_amount decimal,
    start_crypto_value decimal not null,
    end_crypto_value decimal not null,
    selected_end_second int ,
    selected_start_value decimal, 
    selected_end_value decimal,
    won_status varchar not null,
    type varchar not null,
    timestamp timestamp not null,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );

