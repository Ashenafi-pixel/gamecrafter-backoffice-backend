CREATE TABLE users_otp(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT null,
    otp varchar not null,
    created_at TIMESTAMP not null
    );