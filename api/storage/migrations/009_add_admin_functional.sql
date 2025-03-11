ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS moderation_logs (
    id SERIAL PRIMARY KEY,
    action VARCHAR(50) NOT NULL,
    target_id BIGINT NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    performed_by BIGINT NOT NULL REFERENCES users(id),
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_moderation_logs_created_at ON moderation_logs(created_at);
CREATE INDEX idx_moderation_logs_performed_by ON moderation_logs(performed_by);
CREATE INDEX idx_moderation_logs_target_id_type ON moderation_logs(target_id, target_type);