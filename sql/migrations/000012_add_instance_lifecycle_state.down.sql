.ALTER TABLE instances
    DROP COLUMN IF EXISTS readiness_error;

ALTER TABLE instances
    DROP COLUMN IF EXISTS readiness_checked_at;

ALTER TABLE instances
    DROP COLUMN IF EXISTS readiness_status;

ALTER TABLE instances
    DROP COLUMN IF EXISTS lifecycle_state;
