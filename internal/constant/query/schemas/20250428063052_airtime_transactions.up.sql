create table airtime_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID not null,
    transaction_id varchar not null,
    cashout decimal not null,
    billerName varchar not null,
    utilityPackageId int not null,
    packageName varchar not null,
    amount decimal not null,
    status varchar not null,
    timestamp timestamp not null,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)
