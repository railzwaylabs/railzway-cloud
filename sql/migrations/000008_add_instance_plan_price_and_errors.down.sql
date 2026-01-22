ALTER TABLE instances DROP COLUMN IF EXISTS last_error;
ALTER TABLE instances DROP COLUMN IF EXISTS price_id;
ALTER TABLE instances DROP COLUMN IF EXISTS plan_id;

ALTER TABLE organizations
    ALTER COLUMN oss_customer_id TYPE BIGINT USING NULLIF(oss_customer_id, '')::bigint,
    ALTER COLUMN oss_customer_id SET NOT NULL;
