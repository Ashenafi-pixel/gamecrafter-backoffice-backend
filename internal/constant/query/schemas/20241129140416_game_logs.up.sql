CREATE TABLE game_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id UUID not null,
    action VARCHAR(255) NOT null,
    detail JSON ,
    timestamp TIMESTAMP,
    FOREIGN KEY (round_id) REFERENCES rounds(id) ON DELETE CASCADE
);