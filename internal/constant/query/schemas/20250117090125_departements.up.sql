CREATE TABLE departments (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR not null,
   notifications Text[],
   created_at TIMESTAMP
   );