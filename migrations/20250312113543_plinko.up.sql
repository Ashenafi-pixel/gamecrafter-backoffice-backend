CREATE TABLE plinko (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id UUID not null,
 bet_amount decimal not null,
 drop_path varchar not null,
 multiplier decimal,
 win_amount decimal, 
 finalPosition decimal,
 timestamp TIMESTAMP not null DEFAULT now(),
 FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

 );