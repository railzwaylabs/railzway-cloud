ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS lifecycle_state VARCHAR(50) DEFAULT 'serving';

ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS readiness_status VARCHAR(50) DEFAULT 'unknown';

ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS readiness_checked_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS readiness_error TEXT;

UPDATE instances
SET lifecycle_state = 'serving'
WHERE lifecycle_state IS NULL
  AND status IN ('active', 'running');

UPDATE instances
SET lifecycle_state = 'ready'
WHERE lifecycle_state IS NULL
  AND status NOT IN ('active', 'running');
