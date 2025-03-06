CREATE TABLE IF NOT EXISTS interactions (
    id SERIAL PRIMARY KEY,
    entity_type TEXT NOT NULL CHECK(entity_type IN ('joke', 'comment')),
    entity_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(entity_type, entity_id, user_id, type),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
); 