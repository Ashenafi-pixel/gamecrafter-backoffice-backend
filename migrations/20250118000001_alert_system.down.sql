-- Drop alert system tables
DROP TABLE IF EXISTS alert_triggers;
DROP TABLE IF EXISTS alert_configurations;

-- Drop custom types
DROP TYPE IF EXISTS alert_type;
DROP TYPE IF EXISTS alert_status;
DROP TYPE IF EXISTS alert_currency_code;
