-- Add Stripe subscription fields to users table
-- This migration adds the necessary fields for Stripe subscription management
-- Note: SQLite doesn't support adding UNIQUE constraint via ALTER TABLE,
-- so we add columns without UNIQUE and create unique indexes instead

ALTER TABLE users ADD COLUMN stripe_customer_id TEXT DEFAULT NULL;
ALTER TABLE users ADD COLUMN stripe_subscription_id TEXT DEFAULT NULL;
ALTER TABLE users ADD COLUMN subscription_status TEXT DEFAULT NULL;

-- Create unique indexes (SQLite way of enforcing uniqueness on existing table)
CREATE UNIQUE INDEX idx_users_stripe_customer_id ON users(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
CREATE UNIQUE INDEX idx_users_stripe_subscription_id ON users(stripe_subscription_id) WHERE stripe_subscription_id IS NOT NULL;
