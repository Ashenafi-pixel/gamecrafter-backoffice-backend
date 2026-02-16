
create type SpinningWheelMysteryTypes as enum ('point', 'internet_package_in_gb', 'better', 'other');

create table spinning_wheel_mysteries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    Name varchar NOT NULL,
    amount decimal NOT NULL,
    Type SpinningWheelMysteryTypes NOT NULL,
    Status varchar NOT NULL DEFAULT 'ACTIVE',
    frequency int NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users (id),
    deleted_at TIMESTAMP WITH TIME ZONE
)