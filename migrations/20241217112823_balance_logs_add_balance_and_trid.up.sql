ALTER TABLE balance_logs
ADD COLUMN balance_after_update DECIMAL not null DEFAULT 0.0;

ALTER TABLE balance_logs
ADD COLUMN transaction_id VARCHAR NOT NULL DEFAULT '';
