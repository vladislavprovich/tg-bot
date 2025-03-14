CREATE TABLE users (
        user_id UUID PRIMARY KEY,
        telegram_id BIGINT UNIQUE NOT NULL,
        username VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE urls (
        url_id SERIAL PRIMARY KEY,
        user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
        original_url TEXT NOT NULL,
        short_url VARCHAR(255) UNIQUE NOT NULL,
        redirect_count INT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
