-- Migration: Add user_fcm_tokens table for push notifications
-- Up migration

CREATE TABLE IF NOT EXISTS user_fcm_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fcm_token TEXT NOT NULL,
    device VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_user_fcm_token UNIQUE (user_id, fcm_token)
);

-- Index for fast lookups by user
CREATE INDEX idx_fcm_tokens_user_id ON user_fcm_tokens(user_id);

-- Index for finding tokens by branch and role (via join)
CREATE INDEX idx_fcm_tokens_user_branch_role ON user_fcm_tokens(user_id, is_active);
