ALTER TABLE users
ADD COLUMN created_by UUID NULL;

ALTER TABLE users
ADD COLUMN is_admin boolean NULL;