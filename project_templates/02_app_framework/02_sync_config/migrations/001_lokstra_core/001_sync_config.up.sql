-- Migration: Create sync_config table
-- Description: Creates table for storing configuration key-value pairs with JSONB support
-- Author: System
-- Date: 2025-12-02

CREATE SCHEMA IF NOT EXISTS lokstra_auth;

SET SEARCH_PATH TO lokstra_auth;

-- Create sync_config table
CREATE TABLE IF NOT EXISTS sync_config (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on updated_at for faster queries
CREATE INDEX IF NOT EXISTS idx_sync_config_updated_at ON sync_config (updated_at);

-- Create index on value for JSONB queries (optional, uncomment if needed)
-- CREATE INDEX IF NOT EXISTS idx_sync_config_value ON sync_config USING GIN (value);

-- Create trigger function for automatic NOTIFY on changes
CREATE OR REPLACE FUNCTION sync_config_notify()
RETURNS TRIGGER AS $$
DECLARE
    notification JSON;
BEGIN
    -- Build notification payload
    IF (TG_OP = 'DELETE') THEN
        notification = json_build_object(
            'action', 'delete',
            'key', OLD.key,
            'value', null
        );
    ELSE
        notification = json_build_object(
            'action', lower(TG_OP),
            'key', NEW.key,
            'value', NEW.value
        );
    END IF;
    
    -- Send notification to default channel
    -- Channel name can be customized per deployment
    PERFORM pg_notify('config_changes', notification::text);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for INSERT, UPDATE, DELETE
CREATE OR REPLACE TRIGGER sync_config_notify_trigger
    AFTER INSERT OR UPDATE OR DELETE ON sync_config
    FOR EACH ROW
    EXECUTE FUNCTION sync_config_notify();

-- Add comments
COMMENT ON TABLE sync_config IS 'Configuration key-value store with real-time sync support';
COMMENT ON COLUMN sync_config.key IS 'Configuration key (unique identifier)';
COMMENT ON COLUMN sync_config.value IS 'Configuration value stored as JSONB';
COMMENT ON COLUMN sync_config.updated_at IS 'Timestamp of last update';
COMMENT ON FUNCTION sync_config_notify() IS 'Trigger function to send NOTIFY on config changes';
COMMENT ON TRIGGER sync_config_notify_trigger ON sync_config IS 'Automatically sends pg_notify when config changes';
