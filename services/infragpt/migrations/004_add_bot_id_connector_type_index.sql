-- Migration: Add composite index for bot_id and connector_type
-- This migration adds a composite index to optimize the FindIntegrationByBotIDAndType query
-- Run this against the infragpt database

-- Add composite index for bot_id and connector_type
-- This optimizes the FindIntegrationByBotIDAndType query which filters by both fields
CREATE INDEX IF NOT EXISTS idx_integrations_bot_id_connector_type ON integrations (bot_id, connector_type) WHERE bot_id IS NOT NULL;