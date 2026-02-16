BEGIN;
ALTER TABLE lottery_logs DROP COLUMN draw_numbers;
ALTER TABLE lottery_logs ADD COLUMN draw_numbers INTEGER[][];
COMMIT;