create table exchange_rates(
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     currency_from  VARCHAR(3),
     currency_to VARCHAR(3),
     rate decimal, 
     updated_at TIMESTAMP
);