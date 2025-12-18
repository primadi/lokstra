ALTER TABLE users ADD COLUMN phone VARCHAR(20);
ALTER TABLE users ADD COLUMN address TEXT;

CREATE INDEX idx_users_phone ON users(phone);
