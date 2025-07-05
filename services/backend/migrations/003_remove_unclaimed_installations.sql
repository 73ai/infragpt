-- Migration: Remove unclaimed_installations table
-- This migration removes the unclaimed_installations table and related indexes
-- Run this against the infragpt database

-- Drop all indexes first
DROP INDEX IF EXISTS idx_unclaimed_installations_github_id;
DROP INDEX IF EXISTS idx_unclaimed_installations_account;
DROP INDEX IF EXISTS idx_unclaimed_installations_expires;
DROP INDEX IF EXISTS idx_unclaimed_installations_unclaimed;
DROP INDEX IF EXISTS idx_unclaimed_installations_app;
DROP INDEX IF EXISTS idx_unclaimed_installations_active;

-- Drop the table
DROP TABLE IF EXISTS unclaimed_installations;