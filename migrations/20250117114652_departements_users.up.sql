create table departements_users (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   user_id UUID not null,
  department_id UUID not null,
   created_at TIMESTAMPTZ,
   FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
   FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE CASCADE
)