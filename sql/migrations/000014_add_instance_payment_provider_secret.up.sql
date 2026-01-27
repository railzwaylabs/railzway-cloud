ALTER TABLE instances
ADD COLUMN IF NOT EXISTS payment_provider_config_secret TEXT;
