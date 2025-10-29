create table balance_logs(
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id UUID,
     component components not null,
     currency VARCHAR(3),
     change_amount decimal,
     operational_group_id UUID,
     operational_type_id UUID,
     description text, 
     timestamp TIMESTAMP,
     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
     FOREIGN KEY (operational_group_id) REFERENCES operational_groups(id) ON DELETE CASCADE,
     FOREIGN KEY (operational_type_id) REFERENCES operational_types(id) ON DELETE CASCADE
);