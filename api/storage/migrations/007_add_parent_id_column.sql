-- Add parent_id column to comments table for hierarchical comment structure
ALTER TABLE comments ADD COLUMN parent_id BIGINT NULL;

ALTER TABLE comments
    ADD CONSTRAINT fk_comment_parent
        FOREIGN KEY (parent_id)
            REFERENCES comments(id)
            ON DELETE CASCADE;

CREATE INDEX idx_comments_parent_id ON comments(parent_id);