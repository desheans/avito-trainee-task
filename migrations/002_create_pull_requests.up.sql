CREATE TYPE pr_status as ENUM ('MERGED', 'OPEN');

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id TEXT UNIQUE PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id TEXT NOT NULL,
    assigned_reviewers TEXT[],
    status pr_status NOT NULL DEFAULT 'OPEN',
    createdAt TIMESTAMP DEFAULT NOW(),
    mergedAt TIMESTAMP,
    CONSTRAINT fk_author 
        FOREIGN KEY (author_id) REFERENCES users(user_id)
        ON DELETE RESTRICT,
    CONSTRAINT reviewers_len 
        CHECK (array_length(assigned_reviewers, 1) <= 2)
);

CREATE OR REPLACE FUNCTION update_merged_at()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'MERGED' AND OLD.status != 'MERGED' THEN
        NEW.mergedAt = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_merged_at
    BEFORE UPDATE ON pull_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_merged_at();