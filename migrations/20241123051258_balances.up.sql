CREATE TABLE balances (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   user_id UUID not null,
   currency VARCHAR(3) NOT Null,
   real_money decimal,
   bonus_money decimal,
   points INT,
   updated_at TIMESTAMP,
   FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);