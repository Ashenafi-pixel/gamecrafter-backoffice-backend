ALTER TABLE lottery_logs 
ADD COLUMN uniq_identifier uuid NOT NULL DEFAULT gen_random_uuid();