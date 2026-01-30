export interface Organization {
  id: number;
  name: string;
  slug: string;
}

export interface UserProfile {
  email: string;
  first_name?: string;
  last_name?: string;
}

export interface InstanceStatus {
  status: string;
  version: string;
  tier: string;
  plan_id?: string;
  price_id?: string;
  url?: string;
  launch_url?: string;
  last_error?: string;
  subscription_status?: string;
}

export interface TierSpec {
  cpu: string;
  ram: string;
  storage: string;
  isolation: string;
}

export interface PriceMetadata {
  badge: string;
  badge_color: string;
  type: string;
  description: string;
  warning?: string;
  highlight?: string;
  specs: TierSpec;
}

export interface Price {
  id: string;
  product_id: string;
  code: string;
  name: string;
  pricing_model: string;
  billing_mode: string;
  billing_interval: string;
  billing_unit?: string;
  tax_behavior: string;
  active: boolean;
  metadata: PriceMetadata;
}

export interface PriceAmount {
  id: string;
  price_id: string;
  currency: string;
  unit_amount_cents: number;
}
