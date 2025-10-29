create type SpinningWheelTypes as enum ('point', 'internet_package_in_gb', 'better', 'mystery');

create table spinning_wheel_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    Name varchar NOT NULL,
    amount decimal NOT NULL,
    Type SpinningWheelTypes NOT NULL,
    status varchar NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users (id),
    deleted_at TIMESTAMP WITH TIME ZONE,
    Frequency int NOT NULL DEFAULT 1
)