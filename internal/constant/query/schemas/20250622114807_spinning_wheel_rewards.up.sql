create table spinning_wheel_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    amount DECIMAL NOT NULL,
    type VARCHAR(255) NOT NULL,
    status VARCHAR(255) NOT NULL,
    claim_status VARCHAR(255) NOT NULL,
    transaction_id VARCHAR  NULL,
    user_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
)