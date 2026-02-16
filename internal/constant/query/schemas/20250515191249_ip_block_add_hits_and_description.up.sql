ALTER TABLE ip_filters
ADD COLUMN description VARCHAR NOT NULL DEFAULT '';

ALTER TABLE ip_filters ADD COLUMN hits integer not null DEFAULT 0;

ALTER TABLE ip_filters ADD COLUMN last_hit timestamp ;

