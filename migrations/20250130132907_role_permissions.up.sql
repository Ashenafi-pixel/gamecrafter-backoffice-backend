CREATE TABLE role_permissions (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
role_id UUID not null ,
permission_id UUID not null,
FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ;