ALTER TABLE organizations
    ALTER COLUMN oss_customer_id TYPE VARCHAR(255) USING oss_customer_id::text,
    ALTER COLUMN oss_customer_id DROP NOT NULL;

ALTER TABLE instances ADD COLUMN IF NOT EXISTS plan_id VARCHAR(255);
ALTER TABLE instances ADD COLUMN IF NOT EXISTS price_id VARCHAR(255);
ALTER TABLE instances ADD COLUMN IF NOT EXISTS last_error TEXT;
