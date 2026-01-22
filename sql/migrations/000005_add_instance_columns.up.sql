-- Add missing columns to instances table
ALTER TABLE instances ADD COLUMN IF NOT EXISTS tier VARCHAR(50);
ALTER TABLE instances ADD COLUMN IF NOT EXISTS compute_engine VARCHAR(50);
ALTER TABLE instances ADD COLUMN IF NOT EXISTS subscription_id VARCHAR(255);

-- Add index for subscription_id
CREATE INDEX IF NOT EXISTS idx_instances_subscription_id ON instances(subscription_id);
