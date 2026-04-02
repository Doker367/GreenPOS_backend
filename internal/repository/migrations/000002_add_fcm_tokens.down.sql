-- Migration: Add user_fcm_tokens table for push notifications
-- Down migration

DROP TABLE IF EXISTS user_fcm_tokens;
