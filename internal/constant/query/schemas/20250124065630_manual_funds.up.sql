CREATE TABLE manual_funds (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id UUID not null,
 admin_id UUID not null,
 transaction_id varchar not null,
 type varchar not null,
 amount decimal not null,
 currency varchar(3) not null,
 note varchar not null,
 created_at TIMESTAMP not null DEFAULT now(),
 reason varchar not null DEFAULT 'system_restart',
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
FOREIGN KEY (admin_id) REFERENCES users(id) ON DELETE CASCADE
);