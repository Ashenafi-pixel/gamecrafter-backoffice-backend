CREATE TABLE company (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_name VARCHAR(255) NOT NULL,
    support_email VARCHAR(255) NOT NULL CHECK (support_email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    support_phone VARCHAR(50) NOT NULL,
    maintenance_mode BOOLEAN DEFAULT FALSE,
    maximum_login_attempt INTEGER CHECK (maximum_login_attempt > 0),
    password_expiry INTEGER CHECK (password_expiry > 0),
    lockout_duration INTEGER CHECK (lockout_duration > 0),
    require_two_factor_authentication BOOLEAN DEFAULT FALSE,
    ip_list INET[],
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uniq_company_support_phone_active
    ON company(support_phone)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uniq_company_support_phone_deleted
    ON company(support_phone, deleted_at);