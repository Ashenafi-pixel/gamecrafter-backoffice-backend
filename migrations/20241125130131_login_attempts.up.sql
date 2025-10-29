CREATE TABLE login_attempts  ( 
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    ip_address VARCHAR(50) NOT NULL,
    success BOOLEAN NOT NULL,
    attempt_time TIMESTAMP,
    user_agent VARCHAR(50),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    )

