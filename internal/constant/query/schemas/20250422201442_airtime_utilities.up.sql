create table airtime_utilities (
    local_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id int not null,
    productName varchar not null,
    billerName varchar not null,
    amount varchar not null,
    isAmountFixed boolean not null,
    price decimal ,
    status varchar not null,
    timestamp timestamp not null
)
