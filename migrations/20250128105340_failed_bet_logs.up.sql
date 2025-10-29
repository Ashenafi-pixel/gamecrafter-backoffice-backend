CREATE TABLE failed_bet_logs(
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id UUID not null,
     round_id UUID not null,
     bet_id UUID not null,
     manual bool not null DEFAULT true,
     admin_id UUID null,
     status  VARCHAR not null DEFAULT 'IN_PROGRESS',
     created_at TIMESTAMP not null,
     transaction_id UUID not null,
     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
     FOREIGN KEY (round_id) REFERENCES rounds(id) ON DELETE CASCADE,
     FOREIGN KEY (transaction_id) REFERENCES balance_logs(id) ON DELETE CASCADE,
     FOREIGN KEY (bet_id) REFERENCES bets(id) ON DELETE CASCADE
)