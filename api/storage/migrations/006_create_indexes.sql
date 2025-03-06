CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_jokes_author_id ON jokes(author_id);
CREATE INDEX IF NOT EXISTS idx_comments_joke_id ON comments(joke_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_interactions_entity ON interactions(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_interactions_user ON interactions(user_id);
CREATE INDEX IF NOT EXISTS idx_votes_entity ON votes(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_votes_user ON votes(user_id); 