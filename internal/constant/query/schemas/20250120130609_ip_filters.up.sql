create table ip_filters (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   created_by UUID not null,
   start_ip varchar not null,
   end_ip varchar not null,
   type varchar not null,
   created_at TIMESTAMPTZ,
   FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
   )