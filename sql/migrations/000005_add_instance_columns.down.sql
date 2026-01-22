-- Rollback: Remove added columns from instances table
DROP INDEX IF EXISTS idx_instances_subscription_id;

ALTER TABLE instances DROP COLUMN IF EXISTS subscription_id;
ALTER TABLE instances DROP COLUMN IF EXISTS compute_engine;
ALTER TABLE instances DROP COLUMN IF EXISTS tier;
