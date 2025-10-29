CREATE TABLE account_block (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   user_id UUID not null,
   blocked_by UUID not null,
   duration VARCHAR NOT Null,
   type VARCHAR not null,
   blocked_from TIMESTAMPTZ Null,
   blocked_to TIMESTAMPTZ Null,
   unblocked_at TIMESTAMPTZ Null,
   reason VARCHAR Null,
   Note VARCHAR  Null,
   created_at TIMESTAMPTZ,
   FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
   FOREIGN KEY (blocked_by) REFERENCES users(id) ON DELETE CASCADE
);