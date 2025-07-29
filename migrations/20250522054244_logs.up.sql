create table logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID not null,
    module VARCHAR(255) NOT null,
    detail JSON ,
    ip_address VARCHAR(46),
    timestamp TIMESTAMP,
    constraint fk_user foreign key (user_id) references users (id)
);